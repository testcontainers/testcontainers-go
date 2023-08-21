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
		_, _ = rw.Write([]byte(os.Getenv("FOO")))

		rw.WriteHeader(http.StatusAccepted)
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
		panic(err)
	}

	fmt.Println("ready")

	_ = http.Serve(ln, mux)
}
