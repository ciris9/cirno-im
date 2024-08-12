package service

import (
	"cirno-im/logger"
	"cirno-im/services/service/conf"
	"github.com/spf13/viper"
	"os"
	"testing"
)

func TestReadConfig(t *testing.T) {
	viper.SetConfigFile("conf.yaml")
	viper.AddConfigPath("./service")
	logger.Infoln(os.Getwd())
	//viper.AddConfigPath("/etc/conf")
	var config conf.Config
	if err := viper.ReadInConfig(); err != nil {
		t.Error(err)
	} else {
		if err := viper.Unmarshal(&config); err != nil {
			t.Error(err)
		}
	}
	logger.Info(config)
}
