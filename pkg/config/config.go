package config

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Data     DataConfig     `yaml:"data"`
	Registry RegistryConfig `yaml:"registry"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
	HTTP HTTPConfig `yaml:"http"`
}

type HTTPConfig struct {
	Addr    string `yaml:"addr"`
	Timeout int    `yaml:"timeout"`
}

type DataConfig struct {
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	Source string `yaml:"source"`
}

type RedisConfig struct {
	Addr         string `yaml:"addr"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

type RabbitMQConfig struct {
	URL string `yaml:"url"`
}

type RegistryConfig struct {
	Endpoints []string `yaml:"endpoints"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}
