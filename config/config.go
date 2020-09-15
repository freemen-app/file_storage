package config

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/sherifabdlnaby/configuro"
)

type (
	Config struct {
		Api    apiConfig
		Logger loggerConfig
		S3     S3Config
	}

	S3Config struct {
		Bucket string
	}

	apiConfig struct {
		Host string
		Port int
	}

	loggerConfig struct {
		Level  string
		Format string
	}
)

func (c *apiConfig) Addr() string {
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
