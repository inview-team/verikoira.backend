package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type LogLevel string

const (
	Debug LogLevel = "debug"
	Error LogLevel = "error"
	Info  LogLevel = "info"
	Warn  LogLevel = "warn"
)

type DatabaseConfig struct {
	Address    string `json:"address"`
	Database   string `json:"database"`
	Collection string `json:"collection"`
}

type RMQConfig struct {
	Address string `json:"address"`
}

type LoggerSettings struct {
	File  string   `json:"file"`
	Level LogLevel `json:"level"`
}

type Settings struct {
	Host     string         `json:"host"`
	Port     string         `json:"port"`
	Logger   LoggerSettings `json:"logger"`
	Database DatabaseConfig `json:"database"`
	Rmq      RMQConfig      `json:"rabbitmq"`
}

func Init(cfgFile string) (*Settings, error) {
	conf := &Settings{}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		pwd, _ := os.Getwd()
		viper.SetConfigName("configs/example_config")
		viper.AddConfigPath(pwd)
		viper.AutomaticEnv()
		viper.SetConfigType("json")
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file %s: %w", cfgFile, err)
	}

	if err := viper.Unmarshal(conf); err != nil {
		return nil, fmt.Errorf("cannot parse config file %s: %w", cfgFile, err)
	}

	return conf, nil
}
