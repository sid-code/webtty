package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/sid-code/webtty/pkg/sd"
)

type Config struct {
	OneWay         bool   // default false
	Verbose        bool   // default false
	NonInteractive bool   // default true
	StunServer     string // default stun:stun.l.google.com:19302
	Cmd            string // default bash
}

func serve(config Config) {
	conns := make(map[string]hostSession)
	mutex := &sync.RWMutex{}

	http.Handle("/", http.FileServer(http.Dir("./web-client/dist")))

	http.HandleFunc("/init", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			w.WriteHeader(405)
			fmt.Fprintf(w, "Invalid method")
			return
		}
		hs := hostSession{
			oneWay:         config.OneWay,
			cmd:            []string{config.Cmd},
			nonInteractive: config.NonInteractive,
		}
		hs.stunServers = []string{config.StunServer}

		if err := hs.run(); err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Quitting with an unexpected error: \"%s\"\n", err)
		}

		key := sd.Encode(hs.offer)
		mutex.Lock()
		conns[key] = hs
		mutex.Unlock()

		w.WriteHeader(200)
		fmt.Fprintf(w, "%s\n", key)
		return
	})

	http.HandleFunc("/conn", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			w.WriteHeader(405)
			fmt.Fprintf(w, "Invalid method")
			return
		}

		key := req.URL.Query().Get("key")

		mutex.RLock()
		hs, ok := conns[key]
		mutex.RUnlock()

		if !ok {
			w.WriteHeader(404)
			fmt.Fprintf(w, "ERR KEY NOT FOUND\n")
			return
		}
		sdpc, err := io.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "ERR READ BODY %s\n", err)
			return
		}
		sdp, err := sd.Decode(string(sdpc))
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "ERR DECODE %s\n", err)
			return
		}
		hs.answer.Sdp = sdp.Sdp
		logInfo("Answer recieved, connecting...")

		w.WriteHeader(200)
		fmt.Fprintf(w, "SUCCESS")
		go hs.setHostRemoteDescriptionAndWait()

	})
	http.ListenAndServe(":1235", nil)
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
		serve(config)
	} else {
		cc := clientSession{
			offerString: os.Args[2],
		}
		cc.stunServers = []string{config.StunServer}
		err = cc.run()
	}
}
