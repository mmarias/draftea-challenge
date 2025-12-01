package publisher

import (
	"context"
	"errors"
)

type Client interface {
	Publish(ctx context.Context, topic string, message []byte) error
}

type publisher struct {
}

func New() *publisher {
	return &publisher{}
}

func (p *publisher) Publish(ctx context.Context, topic string, message []byte) error {
	if message == nil {
		return errors.New("message cannot be empty")
	}

	return nil
}
