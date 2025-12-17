package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func LoadConfig(name string) (config *Config, err error) {
	dir := GetConfigPath(name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		panic("config path not exist: " + dir)
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
	viper.SetConfigType("env")
	// fmt.Println(path, name)

	// viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	config.DB.Name = config.Project.Name
	config.Project.UploadPath = filepath.Join(config.Project.Root, config.Project.Name, "server", "static")
	config.Project.TemplatesPath = filepath.Join(config.Project.Root, config.Project.Name, "server", "templates")
	return config, err
}
