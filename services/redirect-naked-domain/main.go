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
		// url := RemoveWWWSubdomain(r.URL)
		// http.Redirect(w, r, url.String(), 301)

		fmt.Fprintf(w, "%s %s %s \n", r.Method, r.URL, r.Proto)

		//Iterate over all header fields
		for k, v := range r.Header {
			fmt.Fprintf(w, "Header field %q, Value %q\n", k, v)
		}

		fmt.Fprintf(w, "Host = %q\n", r.Host)
		fmt.Fprintf(w, "RemoteAddr= %q\n", r.RemoteAddr)
		//Get value for a specified token
		fmt.Fprintf(w, "\n\nFinding value of \"Accept\" %q", r.Header["Accept"])
	})

	log.Print("Starting server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
