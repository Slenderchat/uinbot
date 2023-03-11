package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/quotedprintable"
	"net/http"
	"os"
	"regexp"
	"strings"
	"syscall"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/jackc/pgx/v5"
)

var ISUIN int8 = 0
var ISDATA bool = false
var DATA string

var UIN []string
var SUBJECT string
var VEDOMSTVO string
var SUM []string

type Config struct {
	TGtoken    string `json:"tgtoken"`
	TGchatid   int64  `json:"tgchatid"`
	PGpassword string `json:"pgpassword"`
}

func main() {
	var config Config
	configf, err := os.OpenFile("uin.json", os.O_RDONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	jp := json.NewDecoder(configf)
	jp.Decode(&config)
	configf.Close()
	lock, err := os.OpenFile("/tmp/uinbot.lock", os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if ISUIN == 0 {
			r := regexp.MustCompile(`Subject: =\?utf-8\?[bBqQ]\?(.+?)\?=`)
			if r.MatchString(scanner.Text()) {
				SUBJECT = r.ReplaceAllString(scanner.Text(), "$1")
				ISUIN = 1
				continue
			}
		}
		if ISUIN == 1 {
			r := regexp.MustCompile(`\s*=\?utf-8\?[bBqQ]\?(.+?)\?=`)
			if r.MatchString(scanner.Text()) {
				SUBJECT += r.ReplaceAllString(scanner.Text(), "$1")
				continue
			} else {
				ISUIN = 2
			}
		}
		if ISUIN == 2 {
			r := regexp.MustCompile("^=..=..=")
			if r.MatchString(scanner.Text()) {
				d, err := io.ReadAll(quotedprintable.NewReader(strings.NewReader(SUBJECT)))
				if err != nil {
					log.Fatal(err)
				}
				SUBJECT = string(d)
			} else {
				d, err := base64.StdEncoding.DecodeString(SUBJECT)
				if err != nil {
					log.Fatal(err)
				}
				SUBJECT = string(d)
			}
			r = regexp.MustCompile(`УИН\s*по`)
			if r.MatchString(SUBJECT) {
				ISUIN = 3
				continue
			} else {
				log.Fatal("Not uin")
			}
		}
		if !ISDATA {
			r := regexp.MustCompile(`^[\r\n]*$`)
			if r.MatchString(scanner.Text()) {
				ISDATA = true
			}
			continue
		}
		if ISDATA {
			DATA += scanner.Text()
			continue
		}
	}
	if !ISDATA {
		log.Fatal("No data found")
	}
	d, err := base64.StdEncoding.DecodeString(DATA)
	if err != nil {
		log.Fatal(err)
	}
	DATA = string(d)
	r := regexp.MustCompile("[0-9]{20}")
	UIN = r.FindAllString(DATA, -1)
	r = regexp.MustCompile(".*<p>По обращению (.*?) сформирована квитанция.*")
	VEDOMSTVO = r.ReplaceAllString(DATA, "$1")
	r = regexp.MustCompile(`([0-9,]*)\.([0-9]{2}) руб\.`)
	SUM = r.FindAllString(DATA, -1)
	for ind, el := range SUM {
		if r.ReplaceAllString(el, "$2") == "00" {
			el = r.ReplaceAllString(el, "$1")
		} else {
			el = r.ReplaceAllString(el, "$1.$2")
		}
		r1 := regexp.MustCompile(",")
		el = r1.ReplaceAllLiteralString(el, "")
		SUM[ind] = el
	}
	if len(UIN) == 0 {
		log.Fatal("UIN's not found")
	}
	if len(SUM) == 0 {
		log.Fatal("SUM's not found")
	}
	if VEDOMSTVO == "" {
		log.Fatal("VEDOMSTVO not found")
	}
	println("Trying to aquire lock on /tmp/uinbot.lock")
	syscall.Flock(int(lock.Fd()), syscall.LOCK_EX)
	println("Aquired lock on /tmp/uinbot.lock")
	b, err := gotgbot.NewBot(config.TGtoken, &gotgbot.BotOpts{Client: http.Client{}, DefaultRequestOpts: &gotgbot.RequestOpts{Timeout: gotgbot.DefaultTimeout, APIURL: gotgbot.DefaultAPIURL}})
	if err != nil {
		log.Fatal(err)
	}
	pg, err := pgx.Connect(context.Background(), `postgres://uinbot:`+config.PGpassword+`@//uinbot?host=/run/postgresql`)
	if err != nil {
		log.Fatal(err)
	}
	res, err := pg.Query(context.Background(), "SELECT object FROM vedomstvo_objects WHERE vedomstvo = '"+VEDOMSTVO+"'")
	if err != nil {
		log.Fatal(err)
	}
	var objects []string
	for res.Next() {
		var s string
		err := res.Scan(&s)
		if err != nil {
			log.Fatal(err)
		}
		objects = append(objects, s)
	}
	result := "Поступили УИН по следующим объектам:\n"
	if len(objects) == 0 {
		log.Println("WARNING: Linked objects not found for `" + VEDOMSTVO + "`")
		result += "\n1\\. нет информации об объектах\n"
	} else {
		for ind, el := range objects {
			r, err := pg.Query(context.Background(), "SELECT location FROM objects WHERE id = '"+el+"'")
			if err != nil {
				log.Fatal(err)
			}
			var location string
			if r.Next() {
				r.Scan(&location)
				r.Close()
				result += "\n" + fmt.Sprint(ind+1) + "\\. `" + el + "`, адрес: `" + location + "`"
			} else {
				log.Println("WARNING: Location not found for object `" + el + "`")
				result += "\n" + fmt.Sprint(ind+1) + "\\. `" + el + "`, адрес: нет информации об адресе"
			}
			result += ", правообладатели:\n"
			r, err = pg.Query(context.Background(), "SELECT owner FROM object_owners WHERE object = '"+el+"'")
			if err != nil {
				log.Fatal(err)
			}
			var sind int = 1
			for r.Next() {
				var subject string
				r.Scan(&subject)
				result += fmt.Sprint(ind+1) + `\.` + fmt.Sprint(sind) + "\\. `" + subject + "`\n"
				sind++
			}
			if r.CommandTag().RowsAffected() == 0 {
				log.Println("WARNING: No owner found for object `" + el + "`")
				result += fmt.Sprint(ind+1) + `\.` + fmt.Sprint(sind) + "\\. нет информации о правообладателях\n"
			}
		}
	}
	batch := &pgx.Batch{}
	for ind := range UIN {
		batch.Queue("INSERT INTO uins (id, sum) VALUES ('" + UIN[ind] + "', '" + SUM[ind] + "') ON CONFLICT DO NOTHING")
		batch.Queue("INSERT INTO vedomstvo_uins (vedomstvo, uin) SELECT '" + VEDOMSTVO + "', '" + UIN[ind] + "' WHERE NOT EXISTS(SELECT 1 FROM vedomstvo_uins WHERE vedomstvo = '" + VEDOMSTVO + "' AND uin = '" + UIN[ind] + "')")
		result += "\n`" + UIN[ind] + "` на сумму `" + SUM[ind] + "` руб\\.\n"
	}
	pg.SendBatch(context.Background(), batch)
	_, err = b.SendMessage(config.TGchatid, result, &gotgbot.SendMessageOpts{ParseMode: gotgbot.ParseModeMarkdownV2})
	if err != nil {
		log.Fatal(err)
	}
}
