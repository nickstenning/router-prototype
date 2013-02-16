package router

import (
	"net/http"
	"path"
	"sync"
)

type MutableMux struct {
	mu sync.RWMutex
	m  map[string]muxEntry
}

type muxEntry struct {
	explicit bool
	h        http.Handler
}

// NewMutableMux allocates and returns a new MutableMux.
func NewMutableMux() *MutableMux { return &MutableMux{m: make(map[string]muxEntry)} }

// Does path match pattern?
func pathMatch(pattern, path string) bool {
	if len(pattern) == 0 {
		// should not happen
		return false
	}
	n := len(pattern)
	if pattern[n-1] != '/' {
		return pattern == path
	}
	return len(path) >= n && path[0:n] == pattern
}

// Return the canonical path for p, eliminating . and .. elements.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

// Find a handler on a handler map given a path string
// Most-specific (longest) pattern wins
func (mux *MutableMux) match(path string) http.Handler {
	var h http.Handler
	var n = 0
	for k, v := range mux.m {
		if !pathMatch(k, path) {
			continue
		}
		if h == nil || len(k) > n {
			n = len(k)
			h = v.h
		}
	}
	return h
}

// handler returns the handler to use for the request r.
func (mux *MutableMux) handler(r *http.Request) http.Handler {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	// Host-specific pattern takes precedence over generic ones
	h := mux.match(r.Host + r.URL.Path)
	if h == nil {
		h = mux.match(r.URL.Path)
	}
	if h == nil {
		h = http.NotFoundHandler()
	}
	return h
}

// ServeHTTP dispatches the request to the handler whose
// pattern most closely matches the request URL.
func (mux *MutableMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "CONNECT" {
		// Clean path to canonical form and redirect.
		if p := cleanPath(r.URL.Path); p != r.URL.Path {
			w.Header().Set("Location", p)
			w.WriteHeader(http.StatusMovedPermanently)
			return
		}
	}
	mux.handler(r).ServeHTTP(w, r)
}

// Handle registers the handler for the given pattern.
// If a handler already exists for pattern, Handle panics.
func (mux *MutableMux) Handle(pattern string, handler http.Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if pattern == "" {
		panic("router: invalid pattern " + pattern)
	}
	if handler == nil {
		panic("router: nil handler")
	}
	if mux.m[pattern].explicit {
		panic("router: multiple registrations for " + pattern)
	}

	mux.m[pattern] = muxEntry{explicit: true, h: handler}

	// Helpful behavior:
	// If pattern is /tree/, insert an implicit permanent redirect for /tree.
	// It can be overridden by an explicit registration.
	n := len(pattern)
	if n > 0 && pattern[n-1] == '/' && !mux.m[pattern[0:n-1]].explicit {
		mux.m[pattern[0:n-1]] = muxEntry{h: http.RedirectHandler(pattern, http.StatusMovedPermanently)}
	}
}

// HandleFunc registers the handler function for the given pattern.
func (mux *MutableMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	mux.Handle(pattern, http.HandlerFunc(handler))
}

// RemoveHandler removes any registered handlers for the given pattern.
func (mux *MutableMux) RemoveHandler(pattern string) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	_, ok := mux.m[pattern]
	if !ok {
		panic("router: pattern not handled")
	}

	delete(mux.m, pattern)

	// Remove implicit handler if present
	n := len(pattern)
	if n > 0 && pattern[n-1] == '/' && !mux.m[pattern[0:n-1]].explicit {
		delete(mux.m, pattern[0:n-1])
	}
}
