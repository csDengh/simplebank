package utils

import (
	"github.com/spf13/viper"
)

type Config struct {
	ADDR     string `mapstructure:"ADDR"`
	DBDriver string `mapstructure:"DB_Driver"`
	DBSource string `mapstructure:"DB_Source"`
}

func GetConfig(path string) (*Config, error) {
	var config Config

	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return &config, err
	}

	err = viper.Unmarshal(&config)
	return &config, err

}
