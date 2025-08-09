package utils

import (
	"github.com/goccy/go-yaml"
	"os"
	"time"
)

type Config struct {
	BotToken           string        `yaml:"bot_token"`
	DbName             string        `yaml:"db_name"`
	TargetChatId       int64         `yaml:"target_chat_id"`
	AllowedSenderChats []int64       `yaml:"allowed_sender_chats"`
	Interval           time.Duration `yaml:"interval"`
	GroupThreshold     int           `yaml:"group_threshold"`
	WithApi            bool          `yaml:"with_api"`
	ApiKey             string        `yaml:"api_key"`
	ApiPort            int           `yaml:"api_port"`
	UploadChatId       int64         `yaml:"upload_chat_id"`
	Production         bool          `yaml:"production"`
}

func ParseConfig(configFile string) (*Config, error) {
	config := &Config{}
	dat, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(dat, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
