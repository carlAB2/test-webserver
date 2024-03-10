package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/fs/", fileServer)

	go func() {
		fmt.Println("HTTP server is running on :8080")
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Println(err)
		}
	}()

	go func() {
		fmt.Println("HTTPS server is running on :8443")
		err := http.ListenAndServeTLS(":8443", "./certificate/fullchain.pem", "./certificate/privkey.pem", nil)
		if err != nil {
			fmt.Println(err)
		}
	}()

	select {}
}

func fileServer(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path[len("/fs/"):]

	content, err := readFile(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Write(content)
}

func readFile(filePath string) ([]byte, error) {
	content, err := os.ReadFile("./webroot/" + filePath)
	if err != nil {
		return []byte("read " + filePath + " error: " + err.Error()), err
	}
	return content, nil
}
