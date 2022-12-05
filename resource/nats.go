package resource

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

type NatsConnection struct {
	Client *nats.Conn
}

func (conn NatsConnection) Attach() (Resource, error) {
	nc, err := nats.Connect(
		fmt.Sprintf("nats://%s:%d", os.Getenv("NATS_HOST"), 4222),
		nats.PingInterval(20*time.Second),
		nats.MaxPingsOutstanding(5),
		// TODO: this will never stop reconnecting. should it?
		nats.MaxReconnects(-1),
		nats.ReconnectWait(time.Second*5),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			log.Printf("client disconnected: %v", err)
		}),
		nats.ClosedHandler(func(_ *nats.Conn) {
			panic("queue connection closed: goodbye world!")
		}),
		nats.ErrorHandler(func(nc *nats.Conn, s *nats.Subscription, err error) {
			if s != nil {
				log.Printf("Async error in %q/%q: %v", s.Subject, s.Queue, err)
			} else {
				log.Printf("Async error outside subscription: %v", err)
			}
		}))
	if err != nil {
		return nil, err
	}
	conn.Client = nc
	return conn, nil
}
func (conn NatsConnection) Detach() error {
	return nil
}
