package cors

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"strings"
)
// If the Request menthod is not Get, Post, Head, or Options we are returning an error. If it is an option we are changinf the access control request to access control allow.
func cors(w http.ResponseWriter, r *http.Request) (_break bool) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Methods", "GET, HEAD, POST, OPTIONS")

	if !(r.Method == "GET" || r.Method == "POST" || r.Method == "HEAD" || r.Method == "OPTIONS") {
		OutError(w, http.StatusUnauthorized, "Wrong Method")
		return true
	}

	if r.Method == "OPTIONS" {
		for n, h := range r.Header {
			if strings.Contains(n, "Access-Control-Request") {
				for _, h := range h {
					k := strings.Replace(n, "Request", "Allow", 1)
					w.Header().Add(k, h)
				}
			}
		}
		return true
	}
	return
}

func handleHeaders(token string, request *http.Request, r *http.Request) {
	for n, h := range r.Header {
		for _, h := range h {
			request.Header.Add(n, h)
		}
	}
	// TODO change to "taubyte/cors-proxy"
	userAgent := "git/@isomorphic-git/cors-proxy"
	request.Header = r.Header
	request.Header.Set("User-Agent", userAgent)

	tb64 := base64.StdEncoding.EncodeToString([]byte(token))
	request.Header.Set("Authorization", string("Basic ")+tb64)

	// referer
	if r.Header.Get("referer") != "" {
		request.Header.Set("referer", r.Header.Get("referer"))
	}
}
// Here handleResponse function we are thorwing the internal server error when a request is sent to the client and if the response have some error.
// And we are taking the key value pair from the header and if the key is Access-Control-Allow-Origin the we are continuing if not we are creating a loop through the values and adding the actaul value.
func handleResponse(request *http.Request, w http.ResponseWriter) {
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		OutError(w, http.StatusInternalServerError, err.Error())
		return
	}

	defer response.Body.Close()

	// reply
	w.Header().Set("Access-Control-Allow-Origin", "*")
	for k, v := range response.Header {
		if k == "Access-Control-Allow-Origin" {
			continue
		}
		for _, s := range v {
			w.Header().Add(k, s)
		}
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		OutError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(response.StatusCode)

	if _, err := w.Write(body); err != nil {
		log.Printf("write body failed: %v", err)
		return
	}
}

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	if cors(w, r) {
		return
	}

	token := r.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "github ")

	// get URL from request
	URL, err := getURL(r)
	if err != nil {
		OutError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Do request
	request, err := http.NewRequest(r.Method, string(URL), r.Body)
	if err != nil {
		OutError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Handle request headers
	handleHeaders(token, request, r)

	// Handle response
	handleResponse(request, w)
}
