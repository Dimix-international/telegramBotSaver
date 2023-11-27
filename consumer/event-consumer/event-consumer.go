package event_consumer

import (
	"log"
	"sync"
	"telegramBotSaver/events"
	"time"
)

type Consumer struct {
	fetcher events.Fetcher
	processor events.Processor
	batchSize int //сколько событий обработаем за раз
}

func New(fetcher events.Fetcher, processor events.Processor, batchSize int) *Consumer {
	return &Consumer {
		fetcher: fetcher,
		processor: processor,
		batchSize: batchSize,
	}
}

func (c Consumer) Start() error {
	//ждем постоянно события
	for {
		gotEvents, err := c.fetcher.Fetch(c.batchSize)
		if err != nil {
			log.Printf("[err] consumer: %s", err.Error())
			continue
		}

		if len(gotEvents) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		if err := c.handleEvents(gotEvents); err != nil {
			log.Print(err)

			continue
		}
	}
}

//1. Потеря событий:  ретраи, возвращение в храналище, фолбэк, подтверждение для fetcher
//2. обработка всей пачки: останавливаться после первой ошибки, счетчик  ошибок

func (c *Consumer) handleEvents(eventsItems []events.Event) error {
	// for _, event := range eventsItems {
	// 	log.Printf("got new event: %s", event.Text)
    //     //можно подумать о механизме retry, если ошибка
	// 	if err := c.processor.Process(event); err != nil {
	// 		log.Printf("can't handle evenet: %s", err.Error())
	// 		continue
	// 	}
	// }

	// return nil
	var wg sync.WaitGroup
	
	for _, event := range eventsItems {
		wg.Add(1)
		go func(event events.Event) {
			defer wg.Done()	
			log.Printf("got new event: %s", event.Text)
			//можно подумать о механизме retry, если ошибка
			if err := c.processor.Process(event); err != nil {
				log.Printf("can't handle evenet: %s", err.Error())
			}
		} (event)
	}

	wg.Wait()
	return nil
}