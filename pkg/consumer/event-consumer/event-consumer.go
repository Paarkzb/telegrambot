package event_consumer

import (
	"context"
	"log"
	"sync"
	"telegrambot/pkg/events"
	"time"
)

type Consumer struct {
	fetcher   events.Fetcher
	processor events.Processor
	batchSize int
}

func New(fetcher events.Fetcher, processor events.Processor, batchSize int) Consumer {
	return Consumer{
		fetcher:   fetcher,
		processor: processor,
		batchSize: batchSize,
	}
}

func (c Consumer) Start() error {
	for {
		gotEvents, err := c.fetcher.Fetch(c.batchSize)
		if err != nil {
			log.Printf("[ERR] consumer: %s", err.Error())
			continue
		}

		if len(gotEvents) == 0 {
			time.Sleep(1 * time.Second)

			continue
		}

		if err = c.handleEvents(gotEvents); err != nil {
			log.Printf("[ERR] consumer: %s", err.Error())

			continue
		}
	}
}

func (c Consumer) handleEvents(eventsArr []events.Event) error {
	semaphore := make(chan struct{}, 5)

	var wg sync.WaitGroup
	for _, event := range eventsArr {
		wg.Add(1)
		go func(e events.Event) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			log.Printf("got new event: %s", event.Text)

			if err := c.processor.Process(context.TODO(), e); err != nil {
				log.Printf("cant't handle event: %s", err.Error())
				return
			}
		}(event)
	}

	wg.Wait()
	return nil
}
