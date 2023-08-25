package config

import (
	"flag"
	"log"
)

const (
	SourceBinance = "BINANCE"
	SourceUnicorn = "UNICORN"
)

const (
	UnicornAPI = "https://eodhistoricaldata.com/api"
)

type Config struct {
	port         string
	isProduction bool
	unicornKey   string
	isOffline    bool
}

func (c *Config) Port() string {
	return c.port
}

func (c *Config) IsProduction() bool {
	return c.isProduction
}

func (c *Config) UnicornKey() string {
	return c.unicornKey
}

func (c *Config) IsOffline() bool {
	return c.isOffline
}

var serviceConfig = &Config{
	port:         "9701",
	isProduction: true,
	unicornKey:   "",
	isOffline:    false,
}

func ServiceConfig() *Config {
	return serviceConfig
}

func LoadConfig() {

	confPort := flag.String("port", "9701", "port from which to run the service")
	confUnicornKey := flag.String("unicorn-key", "", "Unicorn's EOD API key")
	confIsOffline := flag.Bool("offline", false, "run in offline mode, exchange info will not be up-to-date")
	confIsTestMode := flag.String("mode", "test", "running mode, specify 'prod' to make all symbols available")
	flag.Parse()

	if *confIsTestMode == "prod" {
		serviceConfig.isProduction = true
	} else {
		log.Println("warning: running in test mode")
	}

	if *confIsOffline {
		log.Println("warning: running in offline mode, exchange info will not be up-to-date")
		serviceConfig.isOffline = *confIsOffline
	}

	serviceConfig.port = *confPort
	serviceConfig.unicornKey = *confUnicornKey
}
