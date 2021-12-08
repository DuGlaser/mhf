package mhf

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type Mhf struct {
	server *http.Server
	router *Router
}

type Router struct {
	tree *Node
}

type Node struct {
	parent      *Node
	children    []*Node
	prefix      string
	handler     map[string]http.HandlerFunc
	middlewares []MiddlewareFunc
}

var (
	methods = [...]string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
	}
)

type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

func New() *Mhf {
	m := &Mhf{
		server: new(http.Server),
		router: &Router{
			tree: &Node{
				prefix:      "",
				children:    make([]*Node, 0),
				handler:     make(map[string]http.HandlerFunc),
				middlewares: make([]MiddlewareFunc, 0),
			},
		},
	}

	m.server.Handler = m
	return m
}

func (m *Mhf) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	node, err := m.router.findNode(r.URL.Path)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	handler := node.handler[method]
	if handler == nil {
		w.WriteHeader(404)
		return
	}

	reverse := func(ms []MiddlewareFunc) []MiddlewareFunc {
		for i, j := 0, len(ms)-1; i < j; i, j = i+1, j-1 {
			ms[i], ms[j] = ms[j], ms[i]
		}

		return ms
	}

	middlewares := make([]MiddlewareFunc, 0)

	for {
		middlewares = append(middlewares, reverse(node.middlewares)...)
		node = node.parent
		if node == nil {
			break
		}
	}

	for _, m := range middlewares {
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
	if path[0] == '/' {
		path = path[1:]
	}

	if n, _ := m.router.findNode(path); n != nil && n.handler[method] != nil {
		fmt.Printf("%s(%s) is already resisted.", path, method)
		return
	}

	m.router.add(method, path, handler, middlewares...)
}

func (m *Mhf) Get(path string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	m.add(http.MethodGet, path, handler, middlewares...)
}

func (m *Mhf) Post(path string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	m.add(http.MethodPost, path, handler, middlewares...)
}

func (m *Mhf) Put(path string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	m.add(http.MethodPut, path, handler, middlewares...)
}

func (m *Mhf) Delete(path string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	m.add(http.MethodDelete, path, handler, middlewares...)
}

func (m *Mhf) Middleware(path string, middleware MiddlewareFunc) {
	node, err := m.router.findNode(path)
	if err != nil {
		node, err = m.router.createNode(path)
		if err != nil {
			return
		}
	}
	if err != nil {
		return
	}

	node.middlewares = append(node.middlewares, middleware)
}

func (r *Router) add(method, path string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	node, err := r.findNode(path)
	if err != nil {
		node, err = r.createNode(path)
		if err != nil {
			return
		}
	}

	for _, m := range middlewares {
		handler = m(handler)
	}

	node.handler[method] = handler
}

func (r *Router) createNode(path string) (*Node, error) {
	path = deleteSlushPrefix(path)
	s := strings.Split(path, "/")
	var rest []string

	currentNode := r.tree
	for i, si := range s {
		n := currentNode
		for _, ch := range currentNode.children {
			if ch.prefix == si {
				currentNode = ch
				break
			}
		}

		if n == currentNode {
			rest = s[i:]
			break
		}
	}

	parent := currentNode
	for _, ri := range rest {
		n := &Node{
			parent:      parent,
			children:    make([]*Node, 0),
			prefix:      ri,
			handler:     make(map[string]http.HandlerFunc),
			middlewares: make([]MiddlewareFunc, 0),
		}

		parent.children = append(parent.children, n)

		parent = n
	}

	currentNode = parent
	return currentNode, nil
}

func (r *Router) findNode(path string) (*Node, error) {
	path = deleteSlushPrefix(path)
	currentNode := r.tree

	s := strings.Split(path, "/")
	if s[0] == "" {
		return currentNode, nil
	}

	isNotMatch := false
	for _, si := range s {
		n := currentNode
		for _, ch := range currentNode.children {
			if ch.prefix == si {
				currentNode = ch
				break
			}
		}

		if n == currentNode {
			isNotMatch = true
			break
		}
	}

	if isNotMatch {
		return nil, errors.New(fmt.Sprintf("%s is not found", path))
	}

	return currentNode, nil
}

func deleteSlushPrefix(s string) string {
	if s[0] == '/' {
		s = s[1:]
	}

	return s
}
