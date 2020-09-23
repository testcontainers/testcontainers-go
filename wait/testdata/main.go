package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		data, _ := ioutil.ReadAll(req.Body)
		if bytes.Equal(data, []byte("ping")) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("pong"))
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
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
