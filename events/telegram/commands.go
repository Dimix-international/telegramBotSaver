package telegram

import (
	"context"
	"errors"
	"log"
	"net/url"
	"strings"
	"telegramBotSaver/clients/telegram"
	"telegramBotSaver/lib/e"
	"telegramBotSaver/storage"
)

//типы команд c ключевыми словами
const (
	RndCmd = "/rnd"
	HelpCmd = "/help"
	StartCmd = "/start"
)

//что-то вроде api роутера, по тексту сообщения будем понимать какая команда пришла
func (p *Processor) doCmd(text string, chatID int, username string) error {
	//удаляем пробелы в тексте сообщения
	text = strings.TrimSpace(text)

	//запишем логи
	log.Printf("got new command '%s' from '%s'", text, username)

	// список команды
	// добавление страницы - http://
	// получение рандомной страницы -  /rnd
	// помощь - /help
	//  start - команда отправл автоматически, когда начинается общение - отправляем приветствие и help - /start:

	//проверим является ли сообщение ссылкой - для вызова команды добавление страницы 
	if isAddCmd(text) {
		return p.savePage(chatID, text, username)
	}

	switch text {
	case RndCmd:
		return p.sendRandom(chatID, username)
	case HelpCmd:
		return p.sendHelp(chatID)
	case StartCmd:
		return p.sendHello(chatID)
	default:
		//когда пользователь отправит неизвестную команду или случайный текст
		return p.tg.SendMessage(chatID, msgUnknownCommand)
	}
}

func (p *Processor) savePage(chatID int, pageURL string, username string) (err error) {
	defer func () {err = e.WrapIfErr("can't do command: save page", err)} ()

	senderMessagesTg := NewMessagesSender(chatID, p.tg)

	//страницу которую хотим сохранить
	page := &storage.Page {
		URL: pageURL,
		UserName: username,
	}

	//проверяем существует ли уже
	isExists, err := p.storage.IsExists(context.Background() ,page)
	if err != nil {
		return err
	}

	if isExists {
		return senderMessagesTg(msgAlreadyExists)
	}

	//сохраняем страницу
	if err := p.storage.Save(context.Background() ,page); err != nil {
		return err
	}

	//сообщаем пользователю что сохранили
	if err := senderMessagesTg(msgSaved); err != nil {
		return err
	}
	return nil
}

func (p *Processor) sendRandom(chatID int, username string) (err error) {
	defer func () {err = e.WrapIfErr("can't do command: send random", err)} ()

	senderMessagesTg := NewMessagesSender(chatID, p.tg)

	//ищем случ статью
	page, err := p.storage.PickRandom(context.Background(), username)
	if err != nil{

		if errors.Is(err, storage.ErrNoSavedPages) {
			return senderMessagesTg(msgNoSavedPages)
		}

		return err
	}

	//отправляем ссылку пользователю
	if err := senderMessagesTg(page.URL); err!= nil {
		return err
	}

	//удаляем ссылку


	return p.storage.Remove(context.Background(), page) 
}

func (p *Processor) sendHelp(chatID int) error {
	return p.tg.SendMessage(chatID, msgHelp)
}

func (p *Processor) sendHello(chatID int) error {
	return p.tg.SendMessage(chatID, msgHello)
}

//улучшим отправку сообщения используя замыкания
func NewMessagesSender(chatID int, tg *telegram.Client) func(string) error {
	return func(message string) error {
		return tg.SendMessage(chatID, message)
	}
}

func isAddCmd(text string) bool {
	return isURL(text)
}

func isURL(text string) bool {
	//пропубем распарсить текст, считая что это ссылка (подойдут только с протоколом http/s)
	u, err := url.Parse(text)

	return err == nil && u.Host != ""
}