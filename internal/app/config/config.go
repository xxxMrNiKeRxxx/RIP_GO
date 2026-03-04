package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int
	MinioURL    string
}

func NewConfig() (*Config, error) {
	_ = godotenv.Load()

	configName := "config"
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		cfg := &Config{
			ServiceHost: "0.0.0.0",
			ServicePort: 8080,
			MinioURL:   getMinioURL(),
		}
		logrus.Info("config loaded (defaults)")
		return cfg, nil
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}
	if cfg.ServicePort == 0 {
		cfg.ServicePort = 8080
	}
	if cfg.MinioURL == "" {
		cfg.MinioURL = getMinioURL()
	}

	logrus.Info("config loaded")
	return cfg, nil
}

func getMinioURL() string {
	if url := os.Getenv("MINIO_URL"); url != "" {
		return url
	}
	return "http://localhost:9000/test"
}
