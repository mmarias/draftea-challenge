package publisher

import (
	"context"

	"github.com/mmarias/golearn/internal/infraestructure/eventbus"
)

type Client interface {
	Publish(ctx context.Context, topic string, message []byte) error
}

type publisher struct {
	bus eventbus.Client
}

func New(bus eventbus.Client) Client {
	return &publisher{
		bus: bus,
	}
}

func (p *publisher) Publish(ctx context.Context, topic string, message []byte) error {
	return p.bus.Publish(ctx, topic, message)
}
