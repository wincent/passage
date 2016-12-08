package main

import (
	"encoding/json"
	"github.com/keybase/go-keychain"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"syscall"
)

type Request struct {
	Account string
	Service string
}

var cache map[string][]byte

func main() {
	cache = make(map[string][]byte)
	path := getSockPath()
	syscall.Umask(0077)
	listener, err := net.Listen("unix", path)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Print(err)
				return
			}
			go handleConnection(conn)
		}
	}()
	log.Print("Open and ready for business on UNIX socket at ", path)

	// Need to catch signals in order for `defer`-ed clean-up items to run.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	sig := <-c
	log.Print("Got signal ", sig)
}

func getSockPath() string {
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return user.HomeDir + "/.passage.sock"
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	var request Request
	err := decoder.Decode(&request)
	if err != nil {
		log.Print(err)
		return
	}

	cache_key, err := json.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}
	if cached, exists := cache[string(cache_key)]; exists {
		conn.Write(cached)
		log.Print("Wrote result from cache")
		return
	}

	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassGenericPassword)
	query.SetService(request.Service)
	query.SetAccount(request.Account)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnData(true)
	results, err := keychain.QueryItem(query)
	if err != nil {
		log.Print(err)
	} else if len(results) != 1 {
		log.Print("Item not found")
	} else {
		password := results[0].Data
		conn.Write(password)
		cache[string(cache_key)] = password
		log.Print("Wrote result from keychain")
	}
}
