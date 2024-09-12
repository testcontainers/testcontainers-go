package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

func envHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusAccepted)
		rw.Write([]byte(os.Getenv("FOO"))) //nolint:errcheck // Nothing we can usefully do with the error here.
	}
}

func echoHandler(destination *os.File) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		echo := req.URL.Query()["echo"][0]

		l := log.New(destination, "echo ", 0)
		l.Println(echo)

		rw.WriteHeader(http.StatusAccepted)
	}
}

// a simple server that will echo whatever is in the "echo" parameter to stdout
// in the /stdout endpoint or to stderr in the /stderr endpoint
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/stdout", echoHandler(os.Stdout))
	mux.HandleFunc("/stderr", echoHandler(os.Stderr))
	mux.HandleFunc("/env", envHandler())

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("ready")

	if err := http.Serve(ln, mux); err != nil {
		log.Fatal(err)
	}
}
