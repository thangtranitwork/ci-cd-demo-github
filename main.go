package main

import (
	"fmt"
	"log"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from CI/CD!")
}

func main() {
	http.HandleFunc("/", hello)
	log.Println("Server chạy tại :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
