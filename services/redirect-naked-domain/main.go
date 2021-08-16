package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func RemoveWWWSubdomain(url *url.URL) *url.URL {
	hostname := url.Hostname()
	prefix := hostname[:4]

	if prefix == "www." {
		url.Host = strings.Replace(url.Host, "www.", "", 1)
	}

	return url
}

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host, err := url.Parse(fmt.Sprintf("https://%s%s", r.Host, r.URL))
		if err != nil {
			w.WriteHeader(500)
			return
		}

		target := RemoveWWWSubdomain(host)
		http.Redirect(w, r, target.String(), 301)
	})

	log.Print("Starting server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
