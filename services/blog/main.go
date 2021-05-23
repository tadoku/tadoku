package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
)

//go:embed posts-stub.json
var posts string

//go:embed manual-stub.json
var manual string

func main() {
	http.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, posts)
	})

	http.HandleFunc("/pages/manual", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, manual)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	log.Print("Starting server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
