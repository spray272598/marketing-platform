package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Data   DataConfig   `yaml:"data"`
	Log    LogConfig    `yaml:"log"`
}

type ServerConfig struct {
	Addr    string `yaml:"addr"`
	Name    string `yaml:"name"`
	Timeout int    `yaml:"timeout"`
}

type DataConfig struct {
	MySQL  MySQLConfig  `yaml:"mysql"`
	Redis  RedisConfig  `yaml:"redis"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	Stock  StockConfig  `yaml:"stock"`
}

type MySQLConfig struct {
	DSN string `yaml:"dsn"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type RabbitMQConfig struct {
	URL string `yaml:"url"`
}

type StockConfig struct {
	URL string `yaml:"url"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	applyEnvOverrides(cfg)
	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("SERVER_ADDR"); v != "" {
		cfg.Server.Addr = v
	}
	if v := os.Getenv("MYSQL_DSN"); v != "" {
		cfg.Data.MySQL.DSN = v
	}
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		cfg.Data.Redis.Addr = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Data.Redis.Password = v
	}
	if v := os.Getenv("RABBITMQ_URL"); v != "" {
		cfg.Data.RabbitMQ.URL = v
	}
	if v := os.Getenv("STOCK_URL"); v != "" {
		cfg.Data.Stock.URL = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
}

func (c *Config) GetServerAddr() string {
	if c.Server.Addr != "" {
		return c.Server.Addr
	}
	return ":18091"
}

func (c *Config) GetMySQLDSN(defaultDSN string) string {
	if c.Data.MySQL.DSN != "" {
		return c.Data.MySQL.DSN
	}
	return defaultDSN
}

func (c *Config) GetRedisAddr(defaultAddr string) string {
	if c.Data.Redis.Addr != "" {
		return c.Data.Redis.Addr
	}
	return defaultAddr
}

func (c *Config) GetRabbitMQURL(defaultURL string) string {
	if c.Data.RabbitMQ.URL != "" {
		return c.Data.RabbitMQ.URL
	}
	return defaultURL
}

func (c *Config) GetStockURL(defaultURL string) string {
	if c.Data.Stock.URL != "" {
		return c.Data.Stock.URL
	}
	return defaultURL
}

func (c *Config) GetTimeout() int {
	if c.Server.Timeout > 0 {
		return c.Server.Timeout
	}
	return 10
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}
