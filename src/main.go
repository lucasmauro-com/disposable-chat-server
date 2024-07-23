package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var ctx = context.TODO()
var roomNameRegex = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)
var invalidParamMessage = []byte("Provide `name` and `room` query parameters, with size between 3 and 30 characters. `room` must contain only alphanumeric, hyphen and underscore characters.")
var hub *Hub

func increaseSystemUlimit() {
	if runtime.GOOS == "linux" {
		var rLimit syscall.Rlimit
		if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
			log.Fatalf("failed to get rlimit: %v", err)
		}

		rLimit.Cur = rLimit.Max
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
			log.Fatalf("failed to set rlimit: %v", err)
		}

		log.Info("increased system ulimit")
	}
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	room := strings.TrimSpace(r.URL.Query().Get("room"))
	emptyParam := name == "" || room == ""
	shortParam := len(name) < 3 || len(room) < 3
	largeParam := len(name) > 30 || len(room) > 30
	roomNameHasValidCharacters := roomNameRegex.MatchString(room)
	if emptyParam || shortParam || largeParam || !roomNameHasValidCharacters {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(invalidParamMessage)
		return
	}

	serveWs(name, room, w, r)
}

func main() {
	increaseSystemUlimit()

	godotenv.Load()
	isDebug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if isDebug {
		log.SetLevel(log.DebugLevel)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWS(w, r)
	})

	hub = NewHub()
	go hub.run()

	port := fmt.Sprintf(":%s", os.Getenv("SERVER_PORT"))
	log.Infof("websocket server initialising on port %s", port)

	err := http.ListenAndServe(port, nil)
	log.Fatal(err)
}
