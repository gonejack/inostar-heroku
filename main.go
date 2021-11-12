package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func init() {
	time.Local, _ = time.LoadLocation("Asia/Shanghai")
}

func main() {
	port, exist := os.LookupEnv("PORT")
	if !exist {
		port = "8080"
	}
	_ = http.ListenAndServe(fmt.Sprintf(":%s", port), http.HandlerFunc(handle))
}

func handle(w http.ResponseWriter, r *http.Request) {
	title := r.Header.Get("title")
	fmt.Println(title)
	w.WriteHeader(http.StatusOK)
}
