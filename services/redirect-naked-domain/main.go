package main

import (
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
}
