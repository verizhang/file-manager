package messaging

import "context"

type Handler func(
	ctx context.Context,
	payload []byte,
) error

type Messaging interface {
	Publish(
		ctx context.Context,
		topic string,
		payload []byte,
	) error

	Subscribe(
		ctx context.Context,
		topic string,
		handler Handler,
	) error

	Close() error
}
