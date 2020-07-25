package hcli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
)

type JSObj map[string]interface{}

func (j JSObj) Str(k string) string {
	s, _ := j[k].(string)
	return s
}

func (j JSObj) Obj(k string) JSObj {
	o, _ := j[k].(map[string]interface{})
	return o
}

func (j JSObj) Array(k string) []interface{} {
	a, _ := j[k].([]interface{})
	return a
}

func (j JSObj) String() string {
	b, _ := json.MarshalIndent(j, "", "\t")
	return string(b)
}

type client interface {
	Do(*http.Request) (*http.Response, error)
}

type DumpClient struct {
	Client client
}

func (c DumpClient) Do(req *http.Request) (*http.Response, error) {
	bytes, _ := httputil.DumpRequest(req, true)
	fmt.Println(string(bytes))

	resp, err := c.Client.Do(req)
	if err == nil {
		b, _ := httputil.DumpResponse(resp, true)
		fmt.Println(string(b))
	}
	return resp, err
}
