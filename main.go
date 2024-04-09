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
	http.HandleFunc("/cookie/echo_cookie", echoCookie)
	http.HandleFunc("/cookie/set_cookie", setCookie)
	http.HandleFunc("/cookie/delete_cookie", deleteCookie)

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

/*
*   测试chrome浏览器的cookie行为：
*   测试点1：chrome浏览器发送用户在Application->Cookies中自己添加的cookie
		1. 在chrome浏览器的开发者工具中，在Application->Cookies中添加一个cookie
		2. 然后访问http://swt.isanyier.top:8080/cookie/echo_cookie，查看是否能够获取到刚刚添加的cookie
	测试点2：源站响应Set-Cookie头部，chrome浏览器是否会自动保存cookie
		1. 访问http://swt.isanyier.top:8080/cookie/set_cookie?cookie_str=cookie
		2. 在chrome浏览器的开发者工具中，在Application->Cookies中查看是否有对应的cookie
	测试点3：通过set-cookie头部删除指定cookie
		1. 访问http://swt.isanyier.top:8080/cookie/delete_cookie?cookie_key=key
		2. 在chrome浏览器的开发者工具中，在Application->Cookies中查看是否有key的cookie
*/

func printCookie(w http.ResponseWriter, r *http.Request) {
	cookie := r.Header.Values("cookie")
	fmt.Fprintf(w, "Cookie: %s\n", cookie)
}

func echoCookie(w http.ResponseWriter, r *http.Request) {
	printLog(r)
	printCookie(w, r)
}

func setCookie(w http.ResponseWriter, r *http.Request) {
	printLog(r)
	cookieKey := r.URL.Query().Get("cookie_key")
	setCookie := cookieKey + "=cookie_value; expires = Thu, 01 Jan 2025 00:00:00 GMT"
	w.Header().Set("Set-Cookie", setCookie)
	printCookie(w, r)
}

func deleteCookie(w http.ResponseWriter, r *http.Request) {
	printLog(r)
	cookie_key := r.URL.Query().Get("cookie_key")
	w.Header().Set("Set-Cookie", cookie_key+"=; Max-Age=-1")
	printCookie(w, r)
}
