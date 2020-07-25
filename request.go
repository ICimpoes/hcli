package hcli

import (
	"net/http"
)

type Request struct {
	c client
	*http.Request
	err error

	done bool
	resp Response
}

func (r *Request) Err() error {
	return r.err
}

func (r *Request) exec() Response {
	if r.err != nil {
		return Response{Err: r.err}
	}
	if r.done {
		return r.resp
	}
	r.done = true

	res, err := r.c.Do(r.Request)

	resp := Response{Response: res, Err: err}
	r.resp = resp
	return r.resp
}

func (r Request) chain(f func(Response) *Request) Response {
	return r.exec().Chain(f)
}
