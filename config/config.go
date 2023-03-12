package config

import (
	"encoding/json"
	"log"
	"os"
)

type UINBotConfig struct {
	TGtoken     string `json:"tgtoken"`
	TGuinchatid int64  `json:"tguinchatid"`
	TGextchatid int64  `json:"tgextchatid"`
	PGpassword  string `json:"pgpassword"`
}

func ReadConfig(config *UINBotConfig) {
	configf, err := os.OpenFile("uinbot.json", os.O_RDONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	jp := json.NewDecoder(configf)
	jp.Decode(&config)
	configf.Close()
	if config.PGpassword == "" {
		log.Fatal("Please configure PostgreSQL password")
	}
	if config.TGtoken == "" {
		log.Fatal("Please configure Telegram Bot API token")
	}
}
