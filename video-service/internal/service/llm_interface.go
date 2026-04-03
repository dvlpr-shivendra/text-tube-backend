package service

import "context"

type LLMClient interface {
	Summarize(ctx context.Context, text string) (string, error)
}
