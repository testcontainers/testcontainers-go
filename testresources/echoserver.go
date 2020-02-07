package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func Log(c *gin.Context, destination *os.File) {
	echo := c.Request.URL.Query()["echo"][0]

	l := log.New(destination, "echo ", 0)

	l.Println(echo)

	c.AbortWithStatus(202)

}

// a simple server that will echo whatever is in the "echo" parameter to stdout
// in the /ping endpoint
func main() {
	r := gin.New()
	stop := make(chan bool)

	r.GET("/stdout", func(c *gin.Context) {
		Log(c, os.Stdout)
	})

	r.GET("/stderr", func(c *gin.Context) {
		Log(c, os.Stderr)
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	fmt.Println("ready")

	<-stop
}
