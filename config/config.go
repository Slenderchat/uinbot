package main

import (
	"encoding/json"
	"log"
	"os"
)

type UINBotConfig struct {
	TG struct {
		Token           string
		Uin_chatid      int64
		Rosreest_chatid int64
	}
	PG struct {
		Host     string
		Password string
	}
}

var TGtoken *string = new(string)
var TGuinchatid *int64 = new(int64)
var TGextchatid *int64 = new(int64)
var PGpassword *string = new(string)
var PGhost *string = new(string)

func init() {
	configf, err := os.OpenFile("uinbot.json", os.O_RDONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	jp := json.NewDecoder(configf)
	var config UINBotConfig
	jp.Decode(&config)
	configf.Close()
	if config.PG.Password == "" {
		log.Fatal("Please configure PostgreSQL password")
	}
	if config.PG.Host == "" {
		log.Fatal("Please configure PostgreSQL host or path to directory with UNIX socket")
	}
	if config.TG.Token == "" {
		log.Fatal("Please configure Telegram Bot API token")
	}
	*TGtoken = config.TG.Token
	*TGuinchatid = config.TG.Uin_chatid
	*TGextchatid = config.TG.Rosreest_chatid
	*PGpassword = config.PG.Password
	*PGhost = config.PG.Host
}
