package config

import (
	"os"
	"path/filepath"
	"runtime"

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
	if config.Project.UploadPath == "" {
		config.Project.UploadPath = projectSubdir(config.Project.Root, "static")
	}
	if config.Project.TemplatesPath == "" {
		config.Project.TemplatesPath = projectSubdir(config.Project.Root, "templates")
	}
	return config, err
}

// projectSubdir — путь для загрузок/шаблонов с дефолтами, пригодными в контейнере.
//
// ⚠️ Прежний дефолт клеил `<ROOT>/<NAME>/server/<sub>` без учёта значения `ROOT`
// из прод-`.env` (`/rdkb/`) — получался `//rdkb/food/server/static`, и `mkdir`
// уходил в корень контейнера (`permission denied`, 500 на загрузке фото).
//
// Порядок: `UPLOAD_DIR` → `UPLOAD_PATH` (в `.env` задаётся как `…/static`) →
// `STATIC_DIR`/`APP_DIR` → `/app`. Для шаблонов тот же base, подкаталог — `sub`.
//
// ⚠️ Явный `UPLOAD_PATH` — это УЖЕ каталог статики (pros: `/app/static`, его
// монтируют из папки репозитория), поэтому `sub` к нему не приклеивается:
// `filepath.Join("/app/static", "static")` давал `/app/static/static` — файлы
// уезжали туда, а `GET /api/static/<id>` искал рядом и отдавал 404.
func projectSubdir(root string, sub string) string {
	if base := os.Getenv("UPLOAD_PATH"); base != "" {
		return base
	}
	for _, key := range []string{"UPLOAD_DIR", "STATIC_DIR", "APP_DIR"} {
		if base := os.Getenv(key); base != "" {
			return filepath.Join(base, sub)
		}
	}
	if runtime.GOOS == "linux" {
		return filepath.Join("/app", sub)
	}
	return filepath.Join(root, nameFallback(root), "server", sub)
}

// nameFallback — `NAME` из env, иначе последний сегмент ROOT (локально это имя сервиса).
func nameFallback(root string) string {
	if name := os.Getenv("NAME"); name != "" {
		return name
	}
	clean := filepath.Clean(root)
	return filepath.Base(clean)
}
