package main

import (
	"context"
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
		_, _ = w.Write([]byte("ok"))
	})

	server := http.Server{Addr: "0.0.0.0:6443", Handler: mux}
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
