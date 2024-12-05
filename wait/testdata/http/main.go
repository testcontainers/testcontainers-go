package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/auth-ping", func(w http.ResponseWriter, req *http.Request) {
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
			data, _ := io.ReadAll(req.Body)
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

	mux.HandleFunc("/query-params-ping", func(w http.ResponseWriter, req *http.Request) {
		v := req.URL.Query().Get("v")
		if v == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	})

	mux.HandleFunc("/headers", func(w http.ResponseWriter, req *http.Request) {
		h := req.Header.Get("X-request-header")
		if h != "" {
			w.Header().Add("X-response-header", h)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("headers"))
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	mux.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		data, _ := io.ReadAll(req.Body)
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
		if err := server.ListenAndServeTLS("tls.pem", "tls-key.pem"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println(err)
		}
	}()

	stopsig := make(chan os.Signal, 1)
	signal.Notify(stopsig, syscall.SIGINT, syscall.SIGTERM)
	<-stopsig

	log.Println("stopping...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
