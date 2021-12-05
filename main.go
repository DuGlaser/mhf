package main

import (
	"fmt"
	"net/http"

	"github.com/DuGlaser/mhf/mhf"
)

func main() {
	m := mhf.New()

	m.Get("/foo", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprint(rw, "GET /foo")
	})

	m.Post("/foo", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprint(rw, "POST /foo")
	})

	m.Listen(":8080")
}
