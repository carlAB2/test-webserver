package main

import (
	"fmt"
	"io/ioutil"
	"log"
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

func printLog(r *http.Request) {
	// Print request line, headers, and body to access.log file
	logFile, err := os.OpenFile("access.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening access.log file:", err)
	}
	defer logFile.Close()
	var content = make([]byte, 1024)
	n, err := r.Body.Read(content)
	if err != nil {
		log.Println("Error reading request body:", err)
	}
	log.SetOutput(logFile)
	log.Println("Request Line:", r.Method, r.URL.Path, r.Proto)
	log.Println("Request Headers:")
	for k, v := range r.Header {
		log.Printf("\t%q = %q\n", k, v)
	}
	log.Println("Request Body:", string(content[:n]))

	// Print request line, headers, and body to screen
	fmt.Println("Request Line:", r.Method, r.URL.Path, r.Proto)
	fmt.Println("Request Headers:")
	for k, v := range r.Header {
		fmt.Printf("\t%q = %q\n", k, v)
	}
	fmt.Println("Request Body:", string(content[:n]))
	fmt.Printf("**********************\n")
}

func fileServer(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Path[len("/fs/"):]
	content, err := readFile(filePath)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	printLog(r)
	w.Write(content)
}

func readFile(filePath string) ([]byte, error) {
	content, err := ioutil.ReadFile("./webroot/" + filePath)
	if err != nil {
		return []byte("read " + filePath + " error: " + err.Error()), err
	}
	return content, nil
}
