package config

import (
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

type SymbolSet = map[string]bool

var whitelistLock = sync.Mutex{}
var whitelistCache = map[string]SymbolSet{}
var dataBrokerIdentifier = map[string]string{
	SourceBinance: "binance",
	SourceUnicorn: "unicorn",
}

func loadSymbolList(source string) {
	identifier, ok := dataBrokerIdentifier[source]
	if !ok {
		log.Fatalf("invalid broker %s\n", source)
	}

	testSuffix := ""
	if !ServiceConfig().IsProduction() {
		testSuffix = ".test"
	}

	f, err := os.Open("./assets/" + identifier + testSuffix + ".txt")
	if err != nil {
		log.Fatalf("missing symbol for %s\n", identifier)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("could not read symbol list for %s\n", identifier)
	}

	pairs := strings.Split(string(data), "\n")
	whitelistCache[source] = make(SymbolSet)
	for _, pair := range pairs {
		if len(pair) == 0 {
			continue
		}
		whitelistCache[source][pair] = true
	}

	log.Printf("symbol list for %s loaded, contains %d symbols\n", identifier, len(whitelistCache[source]))
}

func SymbolList(source string) SymbolSet {
	whitelistLock.Lock()
	defer whitelistLock.Unlock()
	if _, ok := whitelistCache[source]; !ok {
		loadSymbolList(source)
	}
	return whitelistCache[source]
}
