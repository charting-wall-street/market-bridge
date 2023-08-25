package main

import (
	"log"
	"marlin/internal/arbiter"
	"marlin/internal/config"
	"marlin/internal/web"
)

func main() {
	log.Println("|- Marlin - Market Linker -|")
	config.LoadConfig()
	arbiter.ExchangeInfo() // preload exchange info
	web.Start()
}
