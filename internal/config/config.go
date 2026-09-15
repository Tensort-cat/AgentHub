package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type MainConfig struct {
	AppName string `toml:"appName"`
	Host    string `toml:"host"`
	Port    int    `toml:"port"`
}

type MysqlConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Database string `toml:"database"`
}

type RedisConfig struct {
	Addr      string `toml:"addr"`
	Password  string `toml:"password"`
	DB        int    `toml:"db"`
	Dimension int    `toml:"dimension"`
}

type LogConfig struct {
	LogPath string `toml:"logPath"`
}

type RabbitmqConfig struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
	Host     string `toml:"host"`
	Port     int    `toml:"prot"`
}

type EmailConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	From     string `toml:"from"`
}

type StaticSrcConfig struct {
	StaticAvatarPath string `toml:"staticAvatarPath"`
	StaticFilePath   string `toml:"staticFilePath"`
}

type Config struct {
	MainConfig      MainConfig      `toml:"mainConfig"`
	MysqlConfig     MysqlConfig     `toml:"mysqlConfig"`
	RedisConfig     RedisConfig     `toml:"redisConfig"`
	LogConfig       LogConfig       `toml:"logConfig"`
	EmailConfig     EmailConfig     `toml:"emailConfig"`
	RabbitmqConfig  RabbitmqConfig  `toml:"rabbitmqConfig"`
	StaticSrcConfig StaticSrcConfig `toml:"staticSrcConfig"`
}

var Cfg *Config

func InitConfig() error {
	cfg := new(Config)
	configPath := os.Getenv("AGENTHUB_CONFIG_PATH")
	if configPath == "" {
		candidates := []string{
			filepath.Join(".", "configs", "config.toml"),
			filepath.Join(".", "config.toml"),
			filepath.Join("D:\\", "dev_soft", "AgentHub", "configs", "config.toml"),
		}
		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				configPath = p
				break
			}
		}
	}
	if configPath == "" {
		configPath = filepath.Join(".", "configs", "config.toml")
	}
	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		log.Fatal(err.Error())
		return err
	}
	Cfg = cfg
	return nil
}
