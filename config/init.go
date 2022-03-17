package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	ServerPort string `mapstructure:"SERVER_PORT"`

	UploadPath    string `mapstructure:"UPLOAD_PATH"`
	TemplatesPath string `mapstructure:"TEMPLATES_PATH"`

	TokenSecret string `mapstructure:"TOKEN_SECRET"`

	RedisHost string `mapstructure:"REDIS_HOST"`
	RedisPort string `mapstructure:"REDIS_PORT"`

	DB            DB            `mapstructure:",squash"`
	Email         Email         `mapstructure:",squash"`
	Social        Social        `mapstructure:",squash"`
	ElasticSearch ElasticSearch `mapstructure:",squash"`
}

type DB struct {
	DB       string `mapstructure:"DB_DB"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	Name     string `mapstructure:"DB_NAME"`
}

type Email struct {
	User     string `mapstructure:"EMAIL_USER"`
	Password string `mapstructure:"EMAIL_PASSWORD"`
	From     string `mapstructure:"EMAIL_FROM"`
	Server   string `mapstructure:"EMAIL_SERVER"`
	Port     string `mapstructure:"EMAIL_PORT"`
}

type Social struct {
	InstagramToken string `mapstructure:"INSTAGRAM_TOKEN"`
	InstagramID    string `mapstructure:"INSTAGRAM_ID"`

	YouTubeApiKey    string `mapstructure:"YOUTUBE_API_KEY"`
	YouTubeChannelID string `mapstructure:"YOUTUBE_CHANNEL_ID"`
}

type ElasticSearch struct {
	ElasticSearchURL string `mapstructure:"ELASTIC_SEARCH_URL"`
	ElasticSearchOn  bool   `mapstructure:"ELASTIC_SEARCH_ON"`
}

func LoadConfig(configPath string) (config *Config, err error) {
	viper.AddConfigPath(configPath)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return config, err
}
