package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		delay := time.Duration(3+rand.Intn(12)) * time.Millisecond
		time.Sleep(delay)
		fmt.Fprintf(w, "ok %s\n", delay)
	})

	fmt.Println("slowweb listening on :8080")
	http.ListenAndServe(":8080", nil)
}
