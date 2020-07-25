package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/icimpoes/hcli"
)

var apiRequest = hcli.New(
	hcli.URL("http://localhost:8000"),
	hcli.Path("api", "v1"),
	SetHeaders,
	hcli.SetClient(hcli.DumpClient{Client: http.DefaultClient}),
)

// custom request builder
func SetHeaders(req *hcli.Request) {
	req.Header.Set("H-1", "1")
	req.Header.Set("H-2", "2")
}

func main() {
	http.Handle("/api/v1/status", http.HandlerFunc(status))
	http.Handle("/api/v1/token", http.HandlerFunc(login))
	http.Handle("/api/v1/user", http.HandlerFunc(user))

	go func() {
		http.ListenAndServe("localhost:8000", nil)
	}()

	for {
		resp := apiRequest.With(
			hcli.Get,
			hcli.Path("status"),
		).Exec()
		if resp.Err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		fmt.Println(resp)
		time.Sleep(10 * time.Millisecond)
	}

	credentials := url.Values{}
	credentials.Set("username", "user1")
	credentials.Set("password", "p@ssw0rd")
	login := apiRequest.With(
		hcli.Path("/token"), hcli.Post, hcli.Form(credentials))

	resp := login.Chain(func(resp hcli.Response) hcli.Builder {
		if resp.StatusCode != http.StatusOK {
			return hcli.SetErr(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
		}
		var body hcli.JSObj
		return hcli.New(hcli.SetErr(resp.JSON(&body)),
			apiRequest.With(hcli.Get, hcli.Path("user"), hcli.Bearer(body.Str("access_token"))))
	}).Exec()

	var body hcli.JSObj
	err := resp.JSON(&body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(body)
}

func status(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func user(w http.ResponseWriter, r *http.Request) {
	if !validHeaders(r.Header) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	toker := r.Header.Get("Authorization")
	if toker != "Bearer 1234" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	b, _ := json.Marshal(hcli.JSObj{"id": 123})
	w.Write(b)
}

func login(w http.ResponseWriter, r *http.Request) {
	if !validHeaders(r.Header) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	if username == "user1" && password == "p@ssw0rd" {
		b, _ := json.Marshal(hcli.JSObj{"access_token": "1234"})
		w.Write(b)
		return
	}

	w.WriteHeader(http.StatusUnauthorized)
}

func validHeaders(header http.Header) bool {
	return header.Get("H-1") == "1" && header.Get("H-2") == "2"
}
