package config

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/viper"
)

func LoadConfig(name string) (config *Config, err error) {
	dir := GetConfigPath(name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		panic("config path not exist: " + dir)
	}

	configFile := path.Join(dir, name+".yaml")
	fmt.Println(configFile)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		panic("config file not exist: " + dir)
	}

	return loadConfig(dir, name)
}

func LoadTestConfig() (config *Config, err error) {
	return LoadConfig("test")
}

func GetConfigPath(name string) string {
	var res string
	if res == "" {
		var err error
		res, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	if res == "" {
		panic("no config path")
	}

	return res
}

func loadConfig(path string, name string) (config *Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(name)
	viper.SetConfigType("yaml")
	fmt.Println(path, name)

	// viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return config, err
}
