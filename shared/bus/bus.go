package bus

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type Bus struct {
	nc *nats.Conn
}

func Connect(url string) (*Bus, error) {
	nc, err := nats.Connect(url, nats.MaxReconnects(-1))
	if err != nil {
		return nil, err
	}
	return &Bus{nc: nc}, nil
}

func (b *Bus) Close() { b.nc.Close() }

func (b *Bus) Publish(subject string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return b.nc.Publish(subject, data)
}

func Subscribe[T any](b *Bus, subject string, handler func(T)) (*nats.Subscription, error) {
	return b.nc.Subscribe(subject, func(msg *nats.Msg) {
		var v T
		if json.Unmarshal(msg.Data, &v) != nil {
			return
		}
		handler(v)
	})
}
