package config

type Config struct {
	Project       Project       `yaml:"project"`
	DB            DB            `yaml:"db"`
	Email         Email         `yaml:"email"`
	Social        Social        `yaml:"social"`
	ElasticSearch ElasticSearch `yaml:"elastic_search"`
	Token         Token         `yaml:"token"`
	Server        Server        `yaml:"server"`
}

type Project struct {
	BinPath       string `yaml:"bin_path"`
	UploadPath    string `yaml:"upload_path"`
	TemplatesPath string `yaml:"templates_path"`
	ModelsPath    string `yaml:"models_path"`
}

type Server struct {
	Port         string `yaml:"port"`
	Host         string `yaml:"host"`
	HTTPS        bool   `yaml:"https"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

type Token struct {
	TokenSecret        string `yaml:"token_secret"`
	TokenAccessMinutes int    `yaml:"token_access_minutes"`
	TokenRefreshHours  int    `yaml:"token_refresh_hours"`
}

type DB struct {
	DB             string `yaml:"db"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	Host           string `yaml:"host"`
	Port           string `yaml:"port"`
	Name           string `yaml:"name"`
	LogPath        string `yaml:"log_path"`
	RemoteUser     string `yaml:"remote_user"`
	RemotePassword string `yaml:"remote_password"`
	Verbose        string `yaml:"verbose"`
}

type Email struct {
	User          string `yaml:"user"`
	Password      string `yaml:"password"`
	From          string `yaml:"from"`
	Server        string `yaml:"server"`
	Port          string `yaml:"port"`
	AuthMethod    string `yaml:"auth_method"`
	WriteTestFile bool   `yaml:"write_test_file"`
}

type Social struct {
	InstagramToken string `yaml:"instagram_token"`
	InstagramID    string `yaml:"instagram_id"`

	YouTubeAPIKey    string `yaml:"youtube_api_key"`
	YouTubeChannelID string `yaml:"youtube_channel_id"`

	VkServiceApplicationKey string `yaml:"vk_service_application_key"`
	VkGroupID               string `yaml:"vk_group_id"`
}

type ElasticSearch struct {
	ElasticSearchURL string `yaml:"elastic_search_url"`
	ElasticSearchOn  bool   `yaml:"elastic_search_on"`
}
