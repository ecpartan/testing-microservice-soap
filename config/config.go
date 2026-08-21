package config

import (
	"flag"
	"log"
	"os"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type MailReport struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	From string `yaml:"from"`
	To   string `yaml:"to"`
	Pass string `yaml:"pass"`
}

type TgReport struct {
	Token string `yaml:"token"`
}

type ServerInfo struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type Config struct {
	ServerInfo ServerInfo `yaml:"server"`
	Mail       MailReport `yaml:"mail"`
	Tg         TgReport   `yaml:"tg"`
}

var once sync.Once
var path string
var config *Config

func LoadConfig() {

	once.Do(func() {
		flag.StringVar(&path, "config", "configs/prod.yaml", "Path to config file")
		flag.Parse()

		if path == "" {
			path = os.Getenv("CONFIG_PATH")
		}
		log.Println(path)

		config = &Config{}

		if err := cleanenv.ReadConfig(path, config); err != nil {
			helptext := "Config file not found. Please create config.yaml file or set CONFIG"
			help, err := cleanenv.GetDescription(config, &helptext)
			log.Println(help)
			log.Fatal(err)
		} else {
			log.Println("Config loaded")
			log.Println(config)
		}
	})

}

func GetMailconfig() *MailReport {
	return &config.Mail
}

func GetTgconfig() *TgReport {
	return &config.Tg
}

func GetServerConfig() *ServerInfo {
	return &config.ServerInfo
}
