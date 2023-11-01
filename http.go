package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/go-cleanhttp"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func serveHTTP(name, addr string) error {
	if !isValidAddr(addr) {
		return fmt.Errorf("-http-addr is invalid %q", addr)
	}
	log.Printf("HTTP listening on %s", addr)

	clientHTTP1 := cleanhttp.DefaultClient()

	clientHTTP2 := cleanhttp.DefaultClient()
	clientHTTP2.Transport = &http2.Transport{
		AllowHTTP: true,
		DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, addr)
		},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/fetch", func(w http.ResponseWriter, r *http.Request) {
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

		var client *http.Client
		switch r.ProtoMajor {
		case 1:
			client = clientHTTP1
		case 2:
			client = clientHTTP2
		default:
			http.Error(w, "unsupported http request protocol version",
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
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

	h2s := &http2.Server{}

	srv := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(mux, h2s),
	}
	return srv.ListenAndServe()
}
