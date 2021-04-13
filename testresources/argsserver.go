package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
)

func envHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {

		rw.Write([]byte(os.Getenv("FOO")))

		rw.WriteHeader(http.StatusAccepted)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/env", envHandler())

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	fmt.Println("ready")

	http.Serve(ln, mux)
}
