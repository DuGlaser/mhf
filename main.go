package main

import (
	"fmt"
	"net/http"

	"github.com/DuGlaser/mhf/mhf"
)

func main() {
	m := mhf.New()

	m.Middleware("/", Logger)
	m.Middleware("/foo", func(hf http.HandlerFunc) http.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request) {
			fmt.Println("/foo")
			hf(rw, r)
		}
	})
	m.Middleware("/foo/bar", func(hf http.HandlerFunc) http.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request) {
			fmt.Println("/bar")
			hf(rw, r)
		}
	})

	m.Get("/foo", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprint(rw, "GET /foo")
	})

	m.Get("/foo/bar", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprint(rw, "GET /foo")
	})

	m.Listen(":8080")
}

func Logger(hf http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s %s\n", r.Proto, r.Method, r.URL.Path)
		hf(rw, r)
	}
}
