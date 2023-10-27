package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func serveHTTP(name, addr string) error {
	if !isValidAddr(addr) {
		return fmt.Errorf("-http-addr is invalid %q", addr)
	}
	log.Printf("HTTP listening on %s", addr)

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
