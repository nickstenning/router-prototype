package router

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Router struct {
	*MutableMux
}

func NewRouter() *Router {
	return &Router{NewMutableMux()}
}

// Add a route to the router
//
// If path ends with a trailing slash, then the route will be considered a
// "prefix" route, and both the path without the trailing slash, and all paths
// under that path, will be added to the router.
//
// e.g.
//
//    r.AddRoute("/foo", "http://google.com") # Only /foo
//
//    r.AddRoute("/bar/", "http://gmail.com") # /bar, /bar/, and /bar/*
//
func (r *Router) AddRoute(path string, dest string) {
	url, err := url.Parse(dest)
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	r.Handle(path, proxy)

	n := len(path)
	if n > 1 && path[n-1] == '/' {
		// Add a prefix route
		r.Handle(path[0:n-1], proxy)
	}
}

// Remove a route from the router
//
// If you are removing a prefix route, you must including the trailing slash
// in path.
//
func (r *Router) RemoveRoute(path string) {
	r.RemoveHandler(path)

	n := len(path)
	if n > 1 && path[n-1] == '/' {
		// Remove the prefix route
		r.RemoveHandler(path[0 : n-1])
	}
}

type RouterApi struct {
	*http.ServeMux
	router *Router
}

func NewRouterApi(router *Router) *RouterApi {
	ret := &RouterApi{http.NewServeMux(), router}

	ret.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		router.AddRoute("/newfoo/", "http://:8081")
		fmt.Fprintf(w, "OK\n")
	})

	ret.HandleFunc("/rmfoo", func(w http.ResponseWriter, r *http.Request) {
		router.RemoveRoute("/newfoo/")
		fmt.Fprintf(w, "OK\n")
	})

	return ret
}
