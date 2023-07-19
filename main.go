package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	OneWay         bool   // default false
	Verbose        bool   // default false
	NonInteractive bool   // default true
	StunServer     string // default stun:stun.l.google.com:19302
	Cmd            string // default bash
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "First argument needs to be a config file\n")
		os.Exit(1)
	}
	configFile := os.Args[1]
	configData, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read config %s: %s\n", configFile, err)
		os.Exit(1)
	}
	var config Config
	_, err = toml.Decode(string(configData), &config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse config %s: %s\n", configFile, err)
		os.Exit(1)
	}

	fmt.Printf("Hello %v\n", config)
	if config.Verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	if len(os.Args) == 2 {
		hc := hostSession{
			oneWay:         config.OneWay,
			cmd:            []string{config.Cmd},
			nonInteractive: config.NonInteractive,
		}
		hc.stunServers = []string{config.StunServer}
		err = hc.run()
	} else {
		cc := clientSession{
			offerString: os.Args[2],
		}
		cc.stunServers = []string{config.StunServer}
		err = cc.run()
	}
	if err != nil {
		fmt.Printf("Quitting with an unexpected error: \"%s\"\n", err)
	}
}
