package appconfig

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server server `yaml:"server"`
	DB     db     `yaml:"db"`
}

type server struct {
	Port string `yaml:"port"`
}

type db struct {
	URL string `yaml:"url"`
}

func GetAppConfig() *Config {
	config := &Config{}
	data, err := os.ReadFile("./configs/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("GetAppConfig:%v", err))
	}
	if err := yaml.Unmarshal(data, config); err != nil {
		panic(fmt.Sprintf("GetAppConfig:%v", err))
	}
	if err := validate(config); err != nil {
		panic(fmt.Sprintf("GetAppConfig:%v", err))
	}
	return config
}

func validate(cfg *Config) error {
	if cfg.DB.URL == "" {
		return fmt.Errorf("appConfig.Validate: empty DB URL ")
	}
	if cfg.Server.Port == "" {
		return fmt.Errorf("appConfig.Validate: empty server port ")
	}

	port, err := strconv.Atoi(cfg.Server.Port)
	if err != nil {
		return fmt.Errorf("appConfig.Validate:%w", err)
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("appConfig.Validate:port out of range (1-65535):%v", port)
	}
	return nil
}
