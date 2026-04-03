package client

import (
	"context"
	"testing"
)

func TestGeminiClient_New(t *testing.T) {
	t.Run("EmptyAPIKey", func(t *testing.T) {
		_, err := NewGeminiClient(context.Background(), "")
		if err == nil {
			t.Fatal("Expected error with empty API key, got nil")
		}
	})
}

func TestGeminiClient_Summarize_EmptyText(t *testing.T) {
	// Note: We can only test the basic logic without external calls here.
	// Since NewGeminiClient checks for API key and the client's Summarize 
	// checks for empty text, we test the latter.
	
	// We can't easily initialize a real client without a real key 
	// for a full unit test, so we'll just test the guards.
	c := &GeminiClient{} // Partially initialized client
	_, err := c.Summarize(context.Background(), "")
	if err == nil {
		t.Fatal("Expected error for empty text, got nil")
	}
}
