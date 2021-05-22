package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
)

//go:embed posts-stub.json
var posts string

func main() {
	http.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, posts)
	})

	log.Print("Starting server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
