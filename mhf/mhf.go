package mhf

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type Mhf struct {
	Router
}

type Router struct {
	tree *Node
}

type Node struct {
	parent         *Node
	children       []*Node
	prefix         string
	handler        map[string]http.HandlerFunc
	middlewares    []MiddlewareFunc
	isVariableNode bool
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
		Router: Router{
			tree: &Node{
				prefix:         "",
				children:       make([]*Node, 0),
				handler:        make(map[string]http.HandlerFunc),
				middlewares:    make([]MiddlewareFunc, 0),
				isVariableNode: false,
			},
		},
	}

	return m
}

func (m *Mhf) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	node, err := m.findNode(r.URL.Path)
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
	http.ListenAndServe(addr, m)
}

func (r *Router) Get(path string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	r.add(http.MethodGet, path, handler, middlewares...)
}

func (r *Router) Post(path string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	r.add(http.MethodPost, path, handler, middlewares...)
}

func (r *Router) Put(path string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	r.add(http.MethodPut, path, handler, middlewares...)
}

func (r *Router) Delete(path string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	r.add(http.MethodDelete, path, handler, middlewares...)
}

func (r *Router) Middleware(path string, middleware MiddlewareFunc) {
	node, err := r.findNode(path)
	if err != nil {
		node, err = r.createNode(path)
		if err != nil {
			return
		}
	}
	if err != nil {
		return
	}

	node.middlewares = append(node.middlewares, middleware)
}

func (r *Router) Group(path string, middlewares ...MiddlewareFunc) (*Router, error) {
	node, err := r.findNode(path)
	if err != nil {
		node, err = r.createNode(path)
		if err != nil {
			return nil, err
		}
	}

	node.middlewares = append(node.middlewares, middlewares...)

	return &Router{
		tree: node,
	}, nil
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
		isVariablePath := false

		if parent != nil && parent.isVariableNode {
			isVariablePath = true
		}

		if len(ri) > 0 && ri[0] == ':' {
			isVariablePath = true
		}

		n := &Node{
			parent:         parent,
			children:       make([]*Node, 0),
			prefix:         ri,
			handler:        make(map[string]http.HandlerFunc),
			middlewares:    make([]MiddlewareFunc, 0),
			isVariableNode: isVariablePath,
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

	matchNodes := []*Node{currentNode}
	for _, si := range s {
		nodes := matchNodes
		matchNodes = []*Node{}

		for _, n := range nodes {
			for _, ch := range n.children {
				if ch.prefix == si {
					matchNodes = append(matchNodes, ch)
				}

				if len(ch.prefix) > 0 && ch.prefix[0] == ':' {
					matchNodes = append(matchNodes, ch)
				}
			}
		}

		if len(matchNodes) == 0 {
			break
		}
	}

	if len(matchNodes) == 0 {
		return nil, errors.New(fmt.Sprintf("%s is not found", path))
	}

	var node *Node
	for _, n := range matchNodes {
		if !n.isVariableNode {
			node = n
			break
		}
	}

	if node == nil {
		node = matchNodes[0]
	}

	return node, nil
}

func deleteSlushPrefix(s string) string {
	if len(s) == 0 {
		return s
	}

	if s[0] == '/' {
		if len(s) == 1 {
			return ""
		}
		s = s[1:]
	}

	return s
}
