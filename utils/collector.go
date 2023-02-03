package utils

import (
	"context"
	"errors"
	"log"

	"github.com/disgoorg/disgo/bot"
)

var (
	ErrTimeout = errors.New("timeout")
)

type CollectedEvent[E bot.Event] struct {
	Data E
}

// Collects the first event of type "E"
// Event must pass filterFunc to be returned, you can create a context with timeout and pass it cancel func to cancel the collector
// Channel is the collected event, returns an error "ErrTimeout" if the timeout is off and a empty struct. Check if empty with reflect.ValueOf(<value>).IsZero()
func NewCollector[E bot.Event](
	client bot.Client,
	ctx context.Context,
	filterFunc func(e E) bool,
	cancelFunc func(),
	ch chan CollectedEvent[E],
	err chan error,
) {
	go func() {
		defer close(ch)

		col, stop := bot.NewEventCollector(client, filterFunc)
	
		defer stop()
	
		select {
			case <- ctx.Done(): {
				cancelFunc()
	
				err <- ErrTimeout
				ch <- CollectedEvent[E]{}
	
				return
			}
			case e := <- col: {
				cancelFunc()
	
				log.Print("it gets the event")
	
				err <- nil
				ch <- CollectedEvent[E]{Data: e}
	
				return
			}
		}
	}()
}