package main

import (
	"context"
	"log"
	"telegrambot/internal/config"
	tgClient "telegrambot/pkg/clients/telegram"
	eventConsumer "telegrambot/pkg/consumer/event-consumer"
	"telegrambot/pkg/events/telegram"
	"telegrambot/pkg/repository/sqlite"
	"telegrambot/pkg/state/redis"
)

func main() {

	cfg := config.MustLoad()
	log.Printf("%v", cfg)

	//rep := files.New(cfg.FilesRepositoryPath)
	rep, err := sqlite.New(cfg.SqliteRepositoryPath)
	if err != nil {
		log.Fatal("can't connect to repository: ", err)
	}

	red := redis.New(cfg)

	err = rep.Init(context.Background())
	if err != nil {
		log.Fatal("can't init repository: ", err)
	}

	eventProcessor := telegram.New(tgClient.New(cfg.TgBotHost, cfg.TgBotToken), rep, red)

	log.Println("service started")

	consumer := eventConsumer.New(eventProcessor, eventProcessor, 100)
	if err := consumer.Start(); err != nil {
		log.Fatal("service is stopped", err)
	}
}
