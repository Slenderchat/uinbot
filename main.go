package main

import (
	"log"
	"os"

	"github.com/Slenderchat/uinbot/uin"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Invalid number of arguments")
	}
	for _, el := range os.Args[1:] {
		if el == "-u" {
			uin.UIN()
		} else {
			log.Fatal("Valid flags are one of: -u or -r")
		}
	}
}
