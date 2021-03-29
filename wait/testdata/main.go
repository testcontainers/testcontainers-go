package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		auth := req.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Basic ") {
			up, err := base64.StdEncoding.DecodeString(auth[6:])
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if string(up) != "admin:admin" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			data, _ := ioutil.ReadAll(req.Body)
			if bytes.Equal(data, []byte("ping")) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("pong"))
				return
			} else {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		w.WriteHeader(http.StatusUnauthorized)
	})

	server := http.Server{Addr: ":6443", Handler: mux}
	go func() {
		log.Println("serving...")
		if err := server.ListenAndServeTLS("tls.pem", "tls-key.pem"); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stopsig := make(chan os.Signal, 1)
	signal.Notify(stopsig, syscall.SIGINT, syscall.SIGTERM)
	<-stopsig

	log.Println("stopping...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_ = server.Shutdown(ctx)
}
