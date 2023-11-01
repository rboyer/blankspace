package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/go-cleanhttp"
)

func serveHTTP(name, addr string) error {
	if !isValidAddr(addr) {
		return fmt.Errorf("-http-addr is invalid %q", addr)
	}
	log.Printf("HTTP listening on %s", addr)

	client := cleanhttp.DefaultClient()

	http.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		u := r.URL.Query().Get("url")
		switch {
		case strings.HasPrefix(u, "http://"):
		case strings.HasPrefix(u, "https://"):
		default:
			http.Error(w, fmt.Sprintf("bad proxy url %q", u),
				http.StatusBadRequest,
			)
			return
		}

		_, err := url.Parse(u)
		if err != nil {
			http.Error(w, fmt.Sprintf("bad proxy url %q: %v", u, err),
				http.StatusBadRequest,
			)
			return
		}

		proxyReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, u, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("bad proxy url %q: %v", u, err),
				http.StatusBadRequest,
			)
			return
		}

		proxyResp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, fmt.Sprintf("error making request: %v", err),
				http.StatusInternalServerError,
			)
			return
		}

		for name := range proxyResp.Header {
			w.Header().Del(name)
			for _, val := range proxyResp.Header.Values(name) {
				w.Header().Add(name, val)
			}
		}

		w.WriteHeader(proxyResp.StatusCode)

		if proxyResp.Body != nil {
			defer proxyResp.Body.Close()
			_, err := io.Copy(w, proxyResp.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("error returning response to caller: %v", err),
					http.StatusInternalServerError,
				)
				return
			}
		}
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		type httpResponse struct {
			Name string
		}

		out := &httpResponse{Name: name}

		enc := json.NewEncoder(w)

		w.Header().Set("content-type", "application/json")
		if err := enc.Encode(&out); err != nil {
			log.Printf("ERROR: %v", err)
		}
	})

	return http.ListenAndServe(addr, nil)
}
