package config

import (
	"fmt"
	"os"
	"path"
	"runtime"

	amqpStore "github.com/freemen-app/amqp-store"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/sherifabdlnaby/configuro"
)

type (
	Config struct {
		Api    ApiConfig
		Logger loggerConfig
		AMQP   amqpStore.Config
		S3     S3Config
	}

	S3Config struct {
		Bucket string
		Region string
	}

	ApiConfig struct {
		Host string
		Port int
	}

	loggerConfig struct {
		Level  string
		Format string
	}
)

func (c ApiConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func New(filename string) *Config {
	_, dir, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	configPath := path.Join(path.Dir(dir), filename)
	if _, err := os.Stat(configPath); err != nil {
		panic(err)
	}
	conf, err := configuro.NewConfig(configuro.WithLoadFromConfigFile(configPath, true))
	if err != nil {
		panic(err)
	}

	confStruct := &Config{}
	if err := conf.Load(confStruct); err != nil {
		panic(err)
	}

	return confStruct
}

func (c *Config) Validate() error {
	return validation.ValidateStruct(
		c,
		validation.Field(&c.Api),
		validation.Field(&c.S3),
		validation.Field(&c.Logger),
	)
}
