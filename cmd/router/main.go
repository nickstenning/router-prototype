package main

import (
	"fmt"
	"net/http"
	"mess/router"
)

func makeTestServer(name string) *http.ServeMux {
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s: %s\n", name, r.URL.Path)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	return mux
}

func main() {
	r := router.NewRouter()

	r.AddRoute("/", "http://:8080/")
	r.AddRoute("/foo/", "http://:8081/")
	r.AddRoute("/bar", "http://:8082/")

	rApi := router.NewRouterApi(r)

	go http.ListenAndServe(":8000", r)
	http.ListenAndServe(":8001", rApi)
}
