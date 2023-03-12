package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"plugin"
	"regexp"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/jackc/pgx/v5"
)

func getConfigItem(p *plugin.Plugin, sname string) plugin.Symbol {
	s, err := p.Lookup(sname)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

func UIN() {
	var isdata bool = false
	var data string
	var uin []string
	var vedomstvo string
	var sum []string
	p, err := plugin.Open("config.so")
	if err != nil {
		log.Fatal(err)
	}
	tgtoken := **getConfigItem(p, "TGtoken").(**string)
	tguinchatid := **getConfigItem(p, "TGuinchatid").(**int64)
	pghost := **getConfigItem(p, "PGhost").(**string)
	pgpassword := **getConfigItem(p, "PGpassword").(**string)
	if tguinchatid == 0 {
		log.Fatal("Please configure Telegram chatid for UIN's")
	}
	scanner := bufio.NewScanner(os.Stdin)
	r := regexp.MustCompile(`^[\r\n]*$`)
	for scanner.Scan() {
		if !isdata {
			if r.MatchString(scanner.Text()) {
				isdata = true
			}
			continue
		}
		if isdata {
			data += scanner.Text()
			continue
		}
	}
	if !isdata {
		log.Fatal("No data found")
	}
	d, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		log.Fatal(err)
	}
	data = string(d)
	r = regexp.MustCompile("[0-9]{20}")
	uin = r.FindAllString(data, -1)
	r = regexp.MustCompile(".*<p>По обращению (.*?) сформирована квитанция.*")
	vedomstvo = r.ReplaceAllString(data, "$1")
	r = regexp.MustCompile(`([0-9,]*)\.([0-9]{2}) руб\.`)
	sum = r.FindAllString(data, -1)
	for ind, el := range sum {
		if r.ReplaceAllString(el, "$2") == "00" {
			el = r.ReplaceAllString(el, "$1")
		} else {
			el = r.ReplaceAllString(el, "$1.$2")
		}
		r1 := regexp.MustCompile(",")
		el = r1.ReplaceAllLiteralString(el, "")
		sum[ind] = el
	}
	if len(uin) == 0 {
		log.Fatal("UIN's not found")
	}
	if len(sum) == 0 {
		log.Fatal("SUM's not found")
	}
	if vedomstvo == "" {
		log.Fatal("VEDOMSTVO not found")
	}
	b, err := gotgbot.NewBot(tgtoken, &gotgbot.BotOpts{Client: http.Client{}, DefaultRequestOpts: &gotgbot.RequestOpts{Timeout: gotgbot.DefaultTimeout, APIURL: gotgbot.DefaultAPIURL}})
	if err != nil {
		log.Fatal(err)
	}
	pg, err := pgx.Connect(context.Background(), `postgres://uinbot:`+pgpassword+`@//uinbot?host=`+pghost)
	if err != nil {
		log.Fatal(err)
	}
	res, err := pg.Query(context.Background(), "SELECT object FROM vedomstvo_objects WHERE vedomstvo = '"+vedomstvo+"'")
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
		log.Println("WARNING: Linked objects not found for `" + vedomstvo + "`")
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
	for ind := range uin {
		batch.Queue("INSERT INTO uins (id, sum) VALUES ('" + uin[ind] + "', '" + sum[ind] + "') ON CONFLICT DO NOTHING")
		batch.Queue("INSERT INTO vedomstvo_uins (vedomstvo, uin) SELECT '" + vedomstvo + "', '" + uin[ind] + "' WHERE NOT EXISTS(SELECT 1 FROM vedomstvo_uins WHERE vedomstvo = '" + vedomstvo + "' AND uin = '" + uin[ind] + "')")
		result += "\n`" + uin[ind] + "` на сумму `" + sum[ind] + "` руб\\.\n"
	}
	pg.SendBatch(context.Background(), batch)
	_, err = b.SendMessage(tguinchatid, result, &gotgbot.SendMessageOpts{ParseMode: gotgbot.ParseModeMarkdownV2})
	if err != nil {
		log.Fatal(err)
	}
}
