package main

import (
	"flag"
	"log"
	"os"
	telegramClient "telegramBotSaver/clients/telegram"
	event_consumer "telegramBotSaver/consumer/event-consumer"
	"telegramBotSaver/events/telegram"
	"telegramBotSaver/storage/files"

	"github.com/joho/godotenv"
)

const (
	storagePath = "files_storage"
	batchSize = 100
)

func main() {
	tgClient := telegramClient.New(mustHost(), mustToken())

	eventsProcessor := telegram.New(tgClient, files.New(storagePath))

	log.Print("..service stared")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal("service stopped")
	}

}

func mustToken() string {
	var token string
	//пробуем сначала получить токен из флага (запускаем  go run main.go --token-bot=123)
	flag.StringVar(&token, "token-bot", "", "token for access to telegram")
	flag.Parse()

	if token != "" {
		return token
	}
	
	godotenv.Load()
	token = os.Getenv("TOKEN_BOT")

	if token == "" {
		log.Fatal("TOKEN is not found")
	}

	return token
}

func mustHost() string {
	var host string
	//пробуем сначала получить токен из флага (запускаем  go run main.go --token-bot=123)
	flag.StringVar(&host, "host-telegram", "", "host for api to telegram")
	flag.Parse()

	if host != "" {
		return host
	}
	
	godotenv.Load()
	host = os.Getenv("TG_BOT_HOST")

	if host == "" {
		log.Fatal("host is not found")
	}

	return host
}