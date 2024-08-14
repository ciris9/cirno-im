package conf

import (
	"fmt"

	"cirno-im/logger"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

// Config Config
type Config struct {
	ServiceID       string   `yaml:"ServiceID"`
	ServiceName     string   `yaml:"ServiceName"`
	Listen          string   `yaml:"Listen" `
	PublicAddress   string   `yaml:"PublicAddress" `
	PublicPort      int      `yaml:"PublicPort"`
	Tags            []string `yaml:"Tags"`
	Domain          string   `yaml:"Domain" `
	ConsulURL       string   `yaml:"ConsulURL"`
	MonitorPort     int      `yaml:"MonitorPort" `
	AppSecret       string   `yaml:"AppSecret"`
	LogLevel        string   `yaml:"LogLevel" `
	MessageGPool    int      `yaml:"MessageGPool" default:"10000"`
	ConnectionGPool int      `yaml:"ConnectionGPool" default:"15000"`
}

// Init InitConfig
func Init(file string) (*Config, error) {
	viper.SetConfigFile(file)
	viper.SetConfigName("conf")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/conf")
	viper.AddConfigPath("F:\\code\\golang\\cirno-im\\services\\gateway")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("conf file not found: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	err := envconfig.Process("", &config)
	if err != nil {
		return nil, err
	}
	logger.Infof("%#v", config)

	return &config, nil
}
