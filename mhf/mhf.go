package mhf

import (
	"errors"
	"fmt"
	"net"
	"net/http"
)

type Mhf struct {
	server *http.Server
	router *Router
}

type Router struct {
	routes map[string]*Route
}

type Route struct {
	method      string
	path        string
	handler     http.HandlerFunc
	middlewares []MiddlewareFunc
}

type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

func New() *Mhf {
	m := &Mhf{
		server: new(http.Server),
		router: &Router{
			routes: make(map[string]*Route),
		},
	}

	m.server.Handler = m
	return m
}

func (m *Mhf) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.URL.Path

	route, err := m.router.find(method, path)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	handler := route.handler
	for _, m := range route.middlewares {
		handler = m(handler)
	}
	handler(w, r)
}

func (m *Mhf) Listen(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	m.server.Serve(l)
}

func (m *Mhf) add(method, path string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	if path[0] != '/' {
		path = fmt.Sprintf("/%s", path)
	}

	if r, _ := m.router.find(method, path); r != nil {
		fmt.Printf("%s(%s) is already resisted.", path, method)
		return
	}

	m.router.add(method, path, handler, middlewares...)
}

func (m *Mhf) Get(path string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	m.add("GET", path, handler, middlewares...)
}

func (m *Mhf) Post(path string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	m.add("POST", path, handler, middlewares...)
}

func (m *Mhf) Put(path string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	m.add("PUT", path, handler, middlewares...)
}

func (m *Mhf) Delete(path string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	m.add("DELETE", path, handler, middlewares...)
}

func (r *Router) add(method, path string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	r.routes[method+path] = &Route{
		method:      method,
		path:        path,
		handler:     handler,
		middlewares: middlewares,
	}
}

func (r *Router) find(method, path string) (*Route, error) {
	rs := r.routes[method+path]
	if rs == nil {
		return nil, errors.New(fmt.Sprintf("%s(%s) is not found", path, method))
	}

	return rs, nil
}
