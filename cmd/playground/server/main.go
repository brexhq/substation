package main

import (
	"net/http"
)

func main() {
	go println("serving at http://localhost:8000/")
	panic(http.ListenAndServe(":8000", http.FileServer(http.Dir("build/playground/"))))
}
