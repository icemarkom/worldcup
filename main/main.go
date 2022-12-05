package main

import (
	"log"
	"os"

	"github.com/icemarkom/worldcup"
)

func main() {
	p := os.Getenv("PORT")
	if p == "" {
		p = "8000"
	}
	if err := worldcup.EntryFunc(p); err != nil {
		log.Fatal(err)
	}
}
