package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func Initialize() error {
	viper.SetConfigName("config")
	viper.SetConfigType("json")

	fpath := os.Getenv("CONFIG_PATH")
	if fpath == "" {
		fpath = "."
	}

	viper.AddConfigPath(fpath)

	viper.SetDefault("nats.urls", []string{"nats://localhost:4222"})
	viper.SetDefault("messages_subscribers_workers_count", 5)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}
	return nil
}

func GetNATSURLs() []string {
	return viper.GetStringSlice("nats.urls")
}

func GetRedisAddress() string {
	return fmt.Sprintf("%s:%d",
		viper.GetString("redis.host"),
		viper.GetInt("redis.port"),
	)
}

func GetMessagesSubscribersWorkersCount() int {
	count := viper.GetInt("nats.messages_subscribers_workers_count")
	if count < 1 {
		return 1
	}
	return count
}
