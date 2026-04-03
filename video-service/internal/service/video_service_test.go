package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "shared/proto"
)

type MockLLMClient struct {
	SummarizeFunc func(ctx context.Context, text string) (string, error)
}

func (m *MockLLMClient) Summarize(ctx context.Context, text string) (string, error) {
	if m.SummarizeFunc != nil {
		return m.SummarizeFunc(ctx, text)
	}
	return "Mock summary", nil
}

func TestSummarizeVideo(t *testing.T) {
	// 1. Mock the transcript service
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		videoId := r.URL.Query().Get("videoId")
		if videoId != "dQw4w9WgXcQ" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"transcript": "This is a test transcript."})
	}))
	defer ts.Close()

	// 2. Mock LLM Client
	mockLLM := &MockLLMClient{
		SummarizeFunc: func(ctx context.Context, text string) (string, error) {
			if text == "This is a test transcript." {
				return "Test Summary Point 1\nTest Summary Point 2", nil
			}
			return "", fmt.Errorf("unexpected text")
		},
	}

	// 3. Initialize VideoService with mocks
	// Note: We don't need real repo or youtube client for SummarizeVideo as it 
	// primarily uses GetVideoTranscript (which we mock via ts.URL) and llmClient.
	svc := &VideoService{
		llmClient:            mockLLM,
		transcriptServiceURL: ts.URL,
	}

	// 4. Test Success Case
	t.Run("Success", func(t *testing.T) {
		req := &pb.SummarizeVideoRequest{
			VideoId: "dQw4w9WgXcQ",
			UserId:  "test-user",
		}
		resp, err := svc.SummarizeVideo(context.Background(), req)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		expectedSummary := "Test Summary Point 1\nTest Summary Point 2"
		if resp.Summary != expectedSummary {
			t.Errorf("Expected summary %q, got %q", expectedSummary, resp.Summary)
		}
		if resp.VideoId != "dQw4w9WgXcQ" {
			t.Errorf("Expected videoId %q, got %q", "dQw4w9WgXcQ", resp.VideoId)
		}
	})

	// 5. Test Failure - Transcript Not Found
	t.Run("TranscriptNotFound", func(t *testing.T) {
		req := &pb.SummarizeVideoRequest{
			VideoId: "invalid-idss", // Still invalid, but should trigger 'invalid' in svc OR not found in ts
			UserId:  "test-user",
		}
		_, err := svc.SummarizeVideo(context.Background(), req)
		if err == nil {
			t.Fatal("Expected error for invalid videoId, got nil")
		}
	})

	// 6. Test Failure - LLM Error
	t.Run("LLMError", func(t *testing.T) {
		mockLLM.SummarizeFunc = func(ctx context.Context, text string) (string, error) {
			return "", fmt.Errorf("llm failed")
		}

		req := &pb.SummarizeVideoRequest{
			VideoId: "dQw4w9WgXcQ",
			UserId:  "test-user",
		}

		_, err := svc.SummarizeVideo(context.Background(), req)
		if err == nil {
			t.Fatal("Expected error for LLM failure, got nil")
		}
	})

	// 7. Test Failure - Empty Transcript
	t.Run("EmptyTranscript", func(t *testing.T) {
		// Create a server that returns empty transcript
		tsEmpty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"transcript": ""})
		}))
		defer tsEmpty.Close()

		svcEmpty := &VideoService{
			llmClient:            mockLLM,
			transcriptServiceURL: tsEmpty.URL,
		}

		req := &pb.SummarizeVideoRequest{
			VideoId: "dQw4w9WgXcQ",
			UserId:  "test-user",
		}
		_, err := svcEmpty.SummarizeVideo(context.Background(), req)
		if err == nil {
			t.Fatal("Expected error for empty transcript, got nil")
		}
	})
}

