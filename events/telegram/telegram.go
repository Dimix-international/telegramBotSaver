package telegram

import (
	"errors"
	"telegramBotSaver/clients/telegram"
	"telegramBotSaver/events"
	"telegramBotSaver/lib/e"
	"telegramBotSaver/storage"
)

type Processor struct {
	tg *telegram.Client
	offset int
	storage storage.Storage
}

//тип Meta, конкретно для telegram
type Meta struct {
	ChatID int
	Username string
}

var ErrUnknownEventType = errors.New("unknown event type")
var ErrUnknownMetaType = errors.New("unknown meta type")

func New(tg *telegram.Client, storage storage.Storage) *Processor {
	return &Processor{
		tg: tg,
		storage: storage,
	}
}

func (p *Processor) Fetch(limit int) ([]events.Event, error){
	updates, err := p.tg.Updates(p.offset, limit)
	if err != nil {
		return nil, e.Wrap("can't get event", err)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates))

	//обходим все update и преобразумем их в event
	for _, u := range updates {
		res = append(res, event(u))
	}

	//обновляем offset,для получ на след запросе новую пачку обновлений
	p.offset = updates[len(updates) - 1].ID + 1

	return res, nil
}

//выполняет действия в зависимости от типа event
func (p *Processor) Process(event events.Event) error {
	switch event.Type {
		case events.Message:
			return p.processMessage(event)
		default: 
			return e.Wrap("can't process message", ErrUnknownEventType)
	}
}

func (p *Processor) processMessage(event events.Event) error  {
	//получаем мету
	meta, err := meta(event)
	if err != nil {
		return e.Wrap("can't process message", err)
	}

	//в зависимости от действия пользователя выполняем нужную команду
	if err := p.doCmd(event.Text, meta.ChatID, meta.Username); err != nil {
		return e.Wrap("can't process message", err)
	}

	return nil
}

func meta(event events.Event) (Meta, error) {
	//делаем type assertion
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("can't get meta", ErrUnknownMetaType)
	}

	return res, nil
}

func event(upd telegram.Update) events.Event {
	updType := fetchType(upd)

	res := events.Event {
		Type: updType,
		Text: fetchText(upd),
	}

	if updType == events.Message {
		//добавляем параметр метад
		res.Meta = Meta{
			ChatID: upd.Message.Chat.ID,
			Username: upd.Message.From.Username,
		}
	}
	return res
}

func fetchType(upd telegram.Update)events.Type {
	if upd.Message == nil {
		return events.Unknown
	}

	return events.Message
}

func fetchText(upd telegram.Update)string {
	if upd.Message == nil {
		return ""
	}

	return upd.Message.Text
}