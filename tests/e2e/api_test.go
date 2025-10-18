package e2e_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/suite"
)

type E2ETestSuite struct {
	suite.Suite
	gatewayURL string
	expect     *httpexpect.Expect
}

func (s *E2ETestSuite) SetupSuite() {
	s.gatewayURL = os.Getenv("GATEWAY_URL")
	if s.gatewayURL == "" {
		s.gatewayURL = "http://localhost:8080"
	}

	s.waitForServices()

	s.expect = httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  s.gatewayURL,
		Reporter: httpexpect.NewAssertReporter(s.T()),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(s.T(), true),
		},
	})
}

func (s *E2ETestSuite) waitForServices() {
	maxRetries := 30
	retryDelay := time.Second

	s.T().Log("Waiting for services to be ready...")

	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(s.gatewayURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			s.T().Log("Services are ready!")
			return
		}
		if resp != nil {
			resp.Body.Close()
		}

		s.T().Logf("Waiting for services... attempt %d/%d", i+1, maxRetries)
		time.Sleep(retryDelay)
	}

	s.T().Fatal("Services failed to start in time")
}

func (s *E2ETestSuite) registerUser(email, password, name string) string {
	resp := s.expect.POST("/api/auth/register").
		WithJSON(map[string]string{
			"email":    email,
			"password": password,
			"name":     name,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	return resp.Value("token").String().Raw()
}

func (s *E2ETestSuite) TestHealthCheck() {
	s.expect.GET("/health").
		Expect().
		Status(http.StatusOK)
}

func (s *E2ETestSuite) TestUserRegistration() {
	timestamp := time.Now().Unix()
	email := fmt.Sprintf("user%d@example.com", timestamp)

	obj := s.expect.POST("/api/auth/register").
		WithJSON(map[string]string{
			"email":    email,
			"password": "SecurePass123!",
			"name":     "Test User",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	obj.Value("token").String().NotEmpty()
	obj.Value("user_id").String().NotEmpty()
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
