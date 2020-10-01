package xconfig

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var (
	envFlag    = flag.String("env", "dev", "environment, e.g. development or production")
	appDirFlag = flag.String("app_dir", "app", "path to the app dir")
)

type Config struct {
	AppDir  string
	Service string
	Env     string

	RedisCache *RedisRing `yaml:"redis_cache"`
	PGMain     *Postgres  `yaml:"pg_main"`

	Uptrace struct {
		DSN string `yaml:"dsn"`
	} `yaml:"uptrace"`

	SecretKey string `yaml:"secret_key"`
}

func LoadConfig(service string) (*Config, error) {
	return loadConfigEnv(service, *appDirFlag, *envFlag)
}

func LoadConfigEnv(service, env string) (*Config, error) {
	return loadConfigEnv(service, *appDirFlag, env)
}

func loadConfigEnv(service, appDir, env string) (*Config, error) {
	appDir, err := filepath.Abs(appDir)
	if err != nil {
		return nil, err
	}

	appDir = findAppDir(appDir, env)
	f, err := os.Open(joinPath(appDir, env))
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	cfg, err := parseConfig(b)
	if err != nil {
		return nil, err
	}

	cfg.AppDir = appDir
	cfg.Service = service
	cfg.Env = env

	return cfg, nil
}

func findAppDir(appDir, env string) string {
	saved := appDir
	for i := 0; i < 10; i++ {
		cfgPath := joinPath(appDir, env)
		_, err := os.Stat(cfgPath)
		if err == nil {
			return appDir
		}

		if appDir == "." {
			break
		}
		appDir = filepath.Dir(filepath.Dir(appDir))
		appDir = filepath.Join(appDir, "app")
	}
	return saved
}

func joinPath(appDir, env string) string {
	return filepath.Join(appDir, "config", env+".yml")
}

func parseConfig(b []byte) (*Config, error) {
	cfg := new(Config)
	err := yaml.Unmarshal(b, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
