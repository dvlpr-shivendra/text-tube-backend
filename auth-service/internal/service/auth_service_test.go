package service

import (
	"context"
	"authservice/internal/models"
	pb "shared/proto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of the UserRepository interface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestRegister(t *testing.T) {
	mockRepo := new(MockUserRepository)
	s := NewAuthService(mockRepo, "secret")

	t.Run("successful registration", func(t *testing.T) {
		req := &pb.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			user := args.Get(1).(*models.User)
			user.ID = primitive.NewObjectID()
		}).Once()

		resp, err := s.Register(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, req.Username, resp.Username)
		assert.NotEmpty(t, resp.Token)
		mockRepo.AssertExpectations(t)
	})

	t.Run("missing fields", func(t *testing.T) {
		req := &pb.RegisterRequest{
			Username: "",
			Email:    "test@example.com",
			Password: "password123",
		}

		resp, err := s.Register(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "missing required fields", err.Error())
	})

	t.Run("user already exists", func(t *testing.T) {
		req := &pb.RegisterRequest{
			Username: "testuser",
			Email:    "existing@example.com",
			Password: "password123",
		}

		existingUser := &models.User{Email: req.Email}
		mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(existingUser, nil).Once()

		resp, err := s.Register(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "user already exists", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("database error on FindByEmail", func(t *testing.T) {
		req := &pb.RegisterRequest{
			Username: "testuser",
			Email:    "error@example.com",
			Password: "password123",
		}

		mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, assert.AnError).Once()

		resp, err := s.Register(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("database error on Create", func(t *testing.T) {
		req := &pb.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError).Once()

		resp, err := s.Register(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, assert.AnError, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestLogin(t *testing.T) {
	mockRepo := new(MockUserRepository)
	s := NewAuthService(mockRepo, "secret")

	t.Run("successful login", func(t *testing.T) {
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		
		user := &models.User{
			ID:       primitive.NewObjectID(),
			Username: "testuser",
			Email:    "test@example.com",
			Password: string(hashedPassword),
		}

		req := &pb.LoginRequest{
			Email:    user.Email,
			Password: password,
		}

		mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(user, nil).Once()

		resp, err := s.Login(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, user.Username, resp.Username)
		assert.Equal(t, user.ID.Hex(), resp.UserId)
		assert.NotEmpty(t, resp.Token)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid credentials - user not found", func(t *testing.T) {
		req := &pb.LoginRequest{
			Email:    "notfound@example.com",
			Password: "password123",
		}

		mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, nil).Once()

		resp, err := s.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "invalid credentials", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid credentials - wrong password", func(t *testing.T) {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
		user := &models.User{
			Email:    "test@example.com",
			Password: string(hashedPassword),
		}

		req := &pb.LoginRequest{
			Email:    user.Email,
			Password: "wrongpassword",
		}

		mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(user, nil).Once()

		resp, err := s.Login(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "invalid credentials", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestValidateToken(t *testing.T) {
	mockRepo := new(MockUserRepository)
	secret := "test-secret"
	s := NewAuthService(mockRepo, secret)

	t.Run("valid token", func(t *testing.T) {
		userID := primitive.NewObjectID().Hex()
		username := "testuser"
		token, _ := s.generateToken(userID, username)

		req := &pb.ValidateTokenRequest{Token: token}
		resp, err := s.ValidateToken(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.Equal(t, userID, resp.UserId)
		assert.Equal(t, username, resp.Username)
	})

	t.Run("invalid token", func(t *testing.T) {
		req := &pb.ValidateTokenRequest{Token: "invalid-token"}
		resp, err := s.ValidateToken(context.Background(), req)

		assert.NoError(t, err)
		assert.False(t, resp.Valid)
	})
}
