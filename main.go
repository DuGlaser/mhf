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

	m.Middleware("/foo", Logger)

	m.Listen(":8080")
}

func Logger(hf http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s %s\n", r.Proto, r.Method, r.URL.Path)
		hf(rw, r)
	}
}
