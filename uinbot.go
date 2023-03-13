package main

import (
	"log"
	"os"
	"plugin"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Invalid number of arguments")
	}
	for _, el := range os.Args[1:] {
		if el == "-u" {
			p, err := plugin.Open("uin.so")
			if err != nil {
				log.Fatal(err)
			}
			uin, err := p.Lookup("UIN")
			if err != nil {
				log.Fatal(err)
			}
			uin.(func())()
		} else if el == "-r" {
			p, err := plugin.Open("rosreestr.so")
			if err != nil {
				log.Fatal(err)
			}
			uin, err := p.Lookup("Rosreestr")
			if err != nil {
				log.Fatal(err)
			}
			uin.(func())()
		} else {
			log.Fatal("Valid flags are one of: -u or -r")
		}
	}
}
