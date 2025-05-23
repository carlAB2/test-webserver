package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	http.HandleFunc("/", root)
	http.HandleFunc("/fs/", fileServer)
	http.HandleFunc("/cookie/echo_cookie", echoCookie)
	http.HandleFunc("/cookie/set_cookie", setCookie)
	http.HandleFunc("/cookie/delete_cookie", deleteCookie)
	http.HandleFunc("/error_code/", handleErrorCode)

	go func() {
		fmt.Println("HTTP server is running on :8080")
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			fmt.Println(err)
		}
	}()

	go func() {
		fmt.Println("HTTPS server is running on :8443")
		err := http.ListenAndServeTLS(":8443", "./certificate/fullchain1.pem", "./certificate/privkey1.pem", nil)
		if err != nil {
			fmt.Println(err)
		}
	}()

	select {}
}

func getClientIP(r *http.Request) string {
    // Check for X-Forwarded-For header first (for clients behind proxy)
    xForwardedFor := r.Header.Get("X-Forwarded-For")
    if xForwardedFor != "" {
        // X-Forwarded-For can contain multiple IPs (client, proxies)
        // The leftmost IP is the original client
        ips := strings.Split(xForwardedFor, ",")
        return strings.TrimSpace(ips[0])
    }
    
    // Check for X-Real-IP header (often added by proxies)
    xRealIP := r.Header.Get("X-Real-IP")
    if xRealIP != "" {
        return xRealIP
    }
    
    // Get IP from RemoteAddr as fallback
    // RemoteAddr is in format "IP:port", so we need to remove the port
    ip := r.RemoteAddr
    if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
        ip = ip[:colonIndex]
    }
    
    // Remove brackets if IPv6
    ip = strings.TrimPrefix(ip, "[")
    ip = strings.TrimSuffix(ip, "]")
    
    return ip
}

func root(w http.ResponseWriter, r *http.Request) {
	printLog(r)
	clientIP := getClientIP(r)
    fmt.Fprintf(w, "Hello, World! Your IP address is: %s", clientIP)
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
		1. 访问http://swt.isanyier.top:8080/cookie/set_cookie?cookie_key=key&path=path&domain=domain
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

func getSetCookieOtherAttrStr(r *http.Request) string {
	setCookieOthAttrStr := ""
	cookiePath := r.URL.Query().Get("path")
	if cookiePath != "" {
		setCookieOthAttrStr += "; path=" + cookiePath
	}
	cookieDomain := r.URL.Query().Get("domain")
	if cookieDomain != "" {
		setCookieOthAttrStr += "; domain=" + cookieDomain
	}
	cookieSameSite := r.URL.Query().Get("samesite")
	if cookieSameSite != "" {
		setCookieOthAttrStr += "; SameSite=" + cookieSameSite
	}
	cookieSecure := r.URL.Query().Get("secure")
	if cookieSecure != "" {
		setCookieOthAttrStr += "; Secure"
	}
	cookieHttpOnly := r.URL.Query().Get("httponly")
	if cookieHttpOnly != "" {
		setCookieOthAttrStr += "; HttpOnly"
	}

	return setCookieOthAttrStr
}

func setCookie(w http.ResponseWriter, r *http.Request) {
	printLog(r)
	cookieKey := r.URL.Query().Get("cookie_key")
	setCookie := cookieKey + "=cookie_value; expires = Thu, 01 Jan 2025 00:00:00 GMT"
	setCookie += getSetCookieOtherAttrStr(r)
	w.Header().Set("Set-Cookie", setCookie)
	printCookie(w, r)
}

func deleteCookie(w http.ResponseWriter, r *http.Request) {
	printLog(r)
	cookie_key := r.URL.Query().Get("cookie_key")
	setCookie := cookie_key + "=; Max-Age=-1"
	setCookie += getSetCookieOtherAttrStr(r)
	w.Header().Set("Set-Cookie", setCookie)
	printCookie(w, r)
}

func handleErrorCode(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) != 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	code, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "Invalid status code", http.StatusBadRequest)
		return
	}

	w.WriteHeader(code)
	fmt.Fprintf(w, "origin server error code %d response page", code)
}
