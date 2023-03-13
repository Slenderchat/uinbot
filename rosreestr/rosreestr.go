package main

import (
	"archive/zip"
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"regexp"
	"strings"
	"syscall"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func getConfigItem(p *plugin.Plugin, sname string) plugin.Symbol {
	s, err := p.Lookup(sname)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

func Rosreestr() {
	var isdata bool
	var data string
	var url string
	var vedomstvo string
	p, err := plugin.Open("config.so")
	if err != nil {
		log.Fatal(err)
	}
	tgtoken := **getConfigItem(p, "TGtoken").(**string)
	tgrosreestrchatid := **getConfigItem(p, "TGrosreestrchatid").(**int64)
	pghost := **getConfigItem(p, "PGhost").(**string)
	pgpassword := **getConfigItem(p, "PGpassword").(**string)
	if tgrosreestrchatid == 0 {
		log.Fatal("Please configure Telegram chatid for rosreestr extractions")
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
	r = regexp.MustCompile(`.*href="(.*?)".*`)
	url = r.ReplaceAllString(data, "$1")
	if len(url) == 0 {
		log.Fatal("URL not found")
	}
	r = regexp.MustCompile(`.*Обработка обращения (.*?) завершена.*`)
	vedomstvo = r.ReplaceAllString(data, "$1")
	if len(vedomstvo) == 0 {
		log.Fatal("Vedomstvo not found")
	}
	b, err := gotgbot.NewBot(tgtoken, &gotgbot.BotOpts{Client: http.Client{}, DefaultRequestOpts: &gotgbot.RequestOpts{Timeout: gotgbot.DefaultTimeout, APIURL: gotgbot.DefaultAPIURL}})
	_ = b
	if err != nil {
		log.Fatal(err)
	}
	_ = pghost
	_ = pgpassword
	/*
		pg, err := pgx.Connect(context.Background(), `postgres://uinbot:`+pgpassword+`@//uinbot?host=`+pghost)
		if err != nil {
			log.Fatal(err)
		}
		_, err = pg.Exec(context.Background(), "INSERT INTO vedomstvo (vedomstvo) VALUES ('"+vedomstvo+"') ON CONFLICT DO NOTHING")
		if err != nil {
			log.Fatal(err)
		}
	*/
	f, err := os.CreateTemp("", "*.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			log.Fatal(err)
		}
		err = os.Remove(f.Name())
		if err != nil {
			log.Fatal(err)
		}
	}()
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		if err != nil {
			log.Fatal(err)
		}
	}()
	clnt := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 YaBrowser/22.11.3.818 Yowser/2.5 Safari/537.36")
	rsp, err := clnt.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer rsp.Body.Close()
	_, err = io.Copy(f, rsp.Body)
	if err != nil {
		log.Fatal(err)
	}
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		log.Fatal(err)
	}
	/*
		defer func() {
			err = os.RemoveAll(dir)
			if err != nil {
				log.Fatal(err)
			}
		}()
	*/
	a, err := zip.OpenReader(f.Name())
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()
	for _, af := range a.File {
		path := filepath.Join(dir, af.Name)
		if af.FileInfo().IsDir() {
			err := os.MkdirAll(path, os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
			continue
		}
		daf, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, af.Mode())
		if err != nil {
			log.Fatal(err)
		}
		defer daf.Close()
		saf, err := af.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer saf.Close()
		_, err = io.Copy(daf, saf)
		if err != nil {
			log.Fatal(err)
		}
	}
	dirls, err := os.ReadDir("uinbot.plugins.d")
	if err != nil {
		log.Fatal(err)
	}
	var plgs []*plugin.Plugin
	for _, el := range dirls {
		if !strings.Contains(el.Name(), ".so") {
			continue
		}
		p, err := plugin.Open(el.Name())
		if err != nil {
			log.Fatal(err)
		}
		plgs = append(plgs, p)
	}
	dirls, err = os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, el := range dirls {
		if strings.Contains(el.Name(), "ЭП") || !strings.Contains(el.Name(), ".xml") || strings.Contains(el.Name(), ".sig") {
			continue
		}
		for _, plg := range plgs {
			s, err := plg.Lookup("Main")
			if err != nil {
				log.Println("WARNING: " + err.Error())
				continue
			}
			s.(func())()
		}
	}
}
