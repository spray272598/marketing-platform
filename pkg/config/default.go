package config

import (
	"os"
)

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Addr:    ":18091",
			Name:    "seckill",
			Timeout: 10,
		},
		Data: DataConfig{
			MySQL: MySQLConfig{
				DSN: "root:root@tcp(127.0.0.1:3306)/marketing_seckill?charset=utf8mb4&parseTime=True&loc=Local",
			},
			Redis: RedisConfig{
				Addr: "127.0.0.1:6379",
			},
			RabbitMQ: RabbitMQConfig{
				URL: "amqp://guest:guest@127.0.0.1:5672/",
			},
			Stock: StockConfig{
				URL: "http://127.0.0.1:18094",
			},
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

func LoadOrDefault(path string) *Config {
	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}
	if path == "" {
		return DefaultConfig()
	}
	cfg, err := Load(path)
	if err != nil {
		return DefaultConfig()
	}
	return cfg
}
