package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	FilesRepositoryPath  string `env:"FILES_REPOSITORY_PATH"`
	SqliteRepositoryPath string `env:"SQLITE_REPOSITORY_PATH"`

	TgBotHost  string `env:"TG_BOT_HOST" `
	TgBotToken string `env:"TG_BOT_TOKEN" `

	RedisAddr     string `env:"REDIS_URL" env-default:"localhost"`
	RedisPort     int    `env:"REDIS_PORT" env-default:"6379"`
	RedisPassword string `env:"REDIS_PASSWORD" env-default:""`
	RedisDB       int    `env:"REDIS_DB" env-default:"0"`
}

func MustLoad() *Config {
	var cfg Config

	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		panic("incorrect env file: " + err.Error())
	}

	return &cfg
}
