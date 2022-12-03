package main

import (
	"os"

	"github.com/icemarkom/worldcup"
)

func main() {
	p := os.Getenv("PORT")
	if p == "" {
		p = "8000"
	}
	worldcup.EntryFunc(p)
}
