package mhf

import (
	"errors"
	"fmt"
	"net"
	"net/http"
)

type Mhf struct {
	server *http.Server
	routes map[string]*Route
}

type Route struct {
	method  string
	path    string
	handler http.HandlerFunc
}

func New() *Mhf {
	m := &Mhf{
		server: new(http.Server),
		routes: make(map[string]*Route),
	}

	m.server.Handler = m
	return m
}

func (m *Mhf) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.URL.Path

	route, err := m.find(method, path)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	route.handler(w, r)
}

func (m *Mhf) Listen(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	m.server.Serve(l)
}

func (m *Mhf) add(method, path string, handler http.HandlerFunc) {
	if m.routes[method+path] != nil {
		fmt.Printf("%s(%s) is already resisted.", path, method)
		return
	}

	m.routes[method+path] = &Route{
		method:  method,
		path:    path,
		handler: handler,
	}
}

func (m *Mhf) find(method, path string) (*Route, error) {
	r := m.routes[method+path]
	if r == nil {
		return nil, errors.New(fmt.Sprintf("%s(%s) is not found", path, method))
	}

	return r, nil
}

func (m *Mhf) Get(path string, handler http.HandlerFunc) {
	m.add("GET", path, handler)
}

func (m *Mhf) Post(path string, handler http.HandlerFunc) {
	m.add("POST", path, handler)
}

func (m *Mhf) Put(path string, handler http.HandlerFunc) {
	m.add("PUT", path, handler)
}

func (m *Mhf) Delete(path string, handler http.HandlerFunc) {
	m.add("DELETE", path, handler)
}
