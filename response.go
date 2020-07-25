package hcli

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	*http.Response
	Err error
}

func (r Response) JSON(v interface{}) error {
	if r.Response != nil && r.Body != nil {
		defer r.Body.Close()
	}
	if r.Err != nil {
		return r.Err
	}
	return json.NewDecoder(r.Response.Body).Decode(v)
}

func (r Response) Chain(f func(Response) *Request) Response {
	if r.Err != nil {
		return r
	}
	return f(r).exec()
}
