package conf

import (
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
)

type Bootstrap struct {
	Server *Server `yaml:"server"`
	Data   *Data   `yaml:"data"`
}

type Server struct {
	HTTP *Server_HTTP `yaml:"http"`
}

type Server_HTTP struct {
	Network string        `yaml:"network"`
	Addr    string        `yaml:"addr"`
	Timeout time.Duration `yaml:"timeout"`
}

type Data struct {
	Database *Data_Database `yaml:"database"`
	Redis    *Data_Redis    `yaml:"redis"`
	RabbitMQ *Data_RabbitMQ `yaml:"rabbitmq"`
	Stock    *Data_Stock    `yaml:"stock"`
}

type Data_Database struct {
	Driver      string `yaml:"driver"`
	Source      string `yaml:"source"`
	Debug       bool   `yaml:"debug"`
	AutoMigrate bool   `yaml:"auto_migrate"`
}

type Data_Redis struct {
	Network      string        `yaml:"network"`
	Addr         string        `yaml:"addr"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"db"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type Data_RabbitMQ struct {
	URL string `yaml:"url"`
}

type Data_Stock struct {
	URL string `yaml:"url"`
}

func (x *Bootstrap) GetServer() *Server {
	if x != nil {
		return x.Server
	}
	return nil
}

func (x *Bootstrap) GetData() *Data {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *Server) GetHttp() *Server_HTTP {
	if x != nil {
		return x.HTTP
	}
	return nil
}

func (x *Server_HTTP) GetNetwork() string {
	if x != nil {
		return x.Network
	}
	return ""
}

func (x *Server_HTTP) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

func (x *Server_HTTP) GetTimeout() time.Duration {
	if x != nil {
		return x.Timeout
	}
	return 0
}

func (x *Data) GetDatabase() *Data_Database {
	if x != nil {
		return x.Database
	}
	return nil
}

func (x *Data) GetRedis() *Data_Redis {
	if x != nil {
		return x.Redis
	}
	return nil
}

func (x *Data) GetRabbitmq() *Data_RabbitMQ {
	if x != nil {
		return x.RabbitMQ
	}
	return nil
}

func (x *Data) GetStock() *Data_Stock {
	if x != nil {
		return x.Stock
	}
	return nil
}

func (x *Data_Database) GetDriver() string {
	if x != nil {
		return x.Driver
	}
	return ""
}

func (x *Data_Database) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *Data_Database) GetDebug() bool {
	if x != nil {
		return x.Debug
	}
	return false
}

func (x *Data_Database) GetAutoMigrate() bool {
	if x != nil {
		return x.AutoMigrate
	}
	return false
}

func (x *Data_Redis) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

func (x *Data_Redis) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *Data_Redis) GetDb() int32 {
	if x != nil {
		return int32(x.DB)
	}
	return 0
}

func (x *Data_Redis) GetReadTimeout() *durationpb.Duration {
	if x != nil {
		return durationpb.New(x.ReadTimeout)
	}
	return nil
}

func (x *Data_Redis) GetWriteTimeout() *durationpb.Duration {
	if x != nil {
		return durationpb.New(x.WriteTimeout)
	}
	return nil
}

func (x *Data_RabbitMQ) GetUrl() string {
	if x != nil {
		return x.URL
	}
	return ""
}

func (x *Data_Stock) GetUrl() string {
	if x != nil {
		return x.URL
	}
	return ""
}
