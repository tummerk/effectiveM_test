package main

import (
	"effectiveMobile_test/internal/app"
	"log"
)

func main() {
	if err := app.Run(); err != nil {
		log.Panicf(`application run error: %v`, err)
	}
}
