package config

type Config struct {
	Project       Project       `mapstructure:",squash" yaml:"project"`
	DB            DB            `mapstructure:",squash" yaml:"db"`
	Email         Email         `mapstructure:",squash" yaml:"email"`
	Social        Social        `mapstructure:",squash" yaml:"social"`
	ElasticSearch ElasticSearch `mapstructure:",squash" yaml:"elastic_search"`
	Token         Token         `mapstructure:",squash" yaml:"token"`
	Server        Server        `mapstructure:",squash" yaml:"server"`
}

type Project struct {
	Name string `mapstructure:"NAME" yaml:"name"`
	Root string `mapstructure:"ROOT" yaml:"root"`

	BinPath    string `mapstructure:"BIN_PATH" yaml:"bin_path"`
	UploadPath string

	TemplatesPath string
	ModelsPath    string `mapstructure:"MODELS_PATH" yaml:"models_path"`
}

type Server struct {
	Port         string `mapstructure:"SERVER_PORT" yaml:"port"`
	Host         string `mapstructure:"SERVER_HOST" yaml:"host"`
	HTTPS        bool   `mapstructure:"SERVER_HTTPS" yaml:"https"`
	ReadTimeout  int    `mapstructure:"SERVER_READ_TIMEOUT" yaml:"read_timeout"`
	WriteTimeout int    `mapstructure:"SERVER_WRITE_TIMEOUT" yaml:"write_timeout"`
}

type Token struct {
	TokenSecret        string `mapstructure:"TOKEN_SECRET" yaml:"token_secret"`
	TokenAccessMinutes int    `mapstructure:"TOKEN_ACCESS_MINUTES" yaml:"token_access_minutes"`
	TokenRefreshHours  int    `mapstructure:"TOKEN_REFRESH_HOURS" yaml:"token_refresh_hours"`
}

type DB struct {
	User     string `mapstructure:"DB_USER" yaml:"user"`
	Password string `mapstructure:"DB_PASSWORD" yaml:"password"`
	Host     string `mapstructure:"DB_HOST" yaml:"host"`
	Port     string `mapstructure:"DB_PORT" yaml:"port"`
	Name     string
	LogPath  string `mapstructure:"DB_LOG_PATH" yaml:"log_path"`
	Verbose  string `mapstructure:"DB_VERBOSE" yaml:"verbose"`
}

type Email struct {
	User          string `mapstructure:"EMAIL_USER" yaml:"user"`
	Password      string `mapstructure:"EMAIL_PASSWORD" yaml:"password"`
	From          string `mapstructure:"EMAIL_FROM" yaml:"from"`
	Server        string `mapstructure:"EMAIL_SERVER" yaml:"server"`
	Port          string `mapstructure:"EMAIL_PORT" yaml:"port"`
	AuthMethod    string `mapstructure:"EMAIL_AUTH_METHOD" yaml:"auth_method"`
	WriteTestFile bool   `mapstructure:"EMAIL_WRITE_TEST_FILE" yaml:"write_test_file"`
}

type Social struct {
	InstagramToken string `mapstructure:"INSTAGRAM_TOKEN" yaml:"instagram_token"`
	InstagramID    string `mapstructure:"INSTAGRAM_ID" yaml:"instagram_id"`

	YouTubeAPIKey    string `mapstructure:"YOUTUBE_API_KEY" yaml:"youtube_api_key"`
	YouTubeChannelID string `mapstructure:"YOUTUBE_CHANNEL_ID" yaml:"youtube_channel_id"`

	VkServiceApplicationKey string `mapstructure:"VK_SERVICE_APPLICATION_KEY" yaml:"vk_service_application_key"`
	VkGroupID               string `mapstructure:"VK_GROUP_ID" yaml:"vk_group_id"`
}

type ElasticSearch struct {
	ElasticSearchURL string `mapstructure:"ELASTIC_SEARCH_URL" yaml:"elastic_search_url"`
	ElasticSearchOn  bool   `mapstructure:"ELASTIC_SEARCH_ON" yaml:"elastic_search_on"`
}
