package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Router struct {
	*http.ServeMux
}

func NewRouter() *Router {
	return &Router{http.NewServeMux()}
}

func (r *Router) AddPrefixRoute(path string, dest string) {
	url, err := url.Parse(dest)
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	r.Handle(path, proxy)
	r.Handle(path + "/", proxy)
}

func (r *Router) AddRoute(path string, dest string) {
	url, err := url.Parse(dest)
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	r.Handle(path, proxy)
}

type RouterApi struct {
	*http.ServeMux
	router *Router
}

func NewRouterApi(router *Router) *RouterApi {
	ret := &RouterApi{http.NewServeMux(), router}

	ret.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		router.AddPrefixRoute("/newfoo", "http://:8081")
		fmt.Fprintf(w, "OK\n");
	})

	return ret
}

func makeTestServer(name string) *http.ServeMux {
	handler := func (w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s: %s\n", name, r.URL.Path)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	return mux
}

func main () {
	// Example application servers
	go http.ListenAndServe(":8080", makeTestServer("fallthrough"))
	go http.ListenAndServe(":8081", makeTestServer("fooPrefix"))
	go http.ListenAndServe(":8082", makeTestServer("bar"))

	router := NewRouter()

	router.AddPrefixRoute("/", "http://:8080/")
	router.AddPrefixRoute("/foo", "http://:8081/")
	router.AddRoute("/bar", "http://:8082/")

	routerApi := NewRouterApi(router)

	go http.ListenAndServe(":8000", router)
	http.ListenAndServe(":8001", routerApi)
}
