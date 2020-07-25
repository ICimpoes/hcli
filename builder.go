package hcli

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

type Builder func(*Request)

func New(bs ...Builder) Builder {
	return MultiBuilder(bs...)
}

func MultiBuilder(bs ...Builder) Builder {
	return func(r *Request) {
		for _, b := range bs {
			if r.err != nil {
				return
			}
			b(r)
		}
	}
}

func (b Builder) With(bs ...Builder) Builder {
	return New(b, New(bs...))
}

func (b Builder) Chain(f func(Response) Builder) Builder {
	return func(req *Request) {
		rsp := b.Exec()
		if rsp.Err != nil {
			SetErr(rsp.Err)(req)
		}
		f(rsp)(req)
	}
}

//func (b Builder) ChainMulti(f func(Response) []Builder) func() {
//	return func() {
//		rsp := b.Exec()
//		if rsp.Err != nil {
//			return
//		}
//		bs := f(rsp)
//		for _, b := range bs {
//			b.Exec()
//		}
//	}
//}

func (b Builder) Exec() Response {
	httpReq, err := http.NewRequest("", "", nil)
	req := &Request{
		c:       http.DefaultClient,
		Request: httpReq,
		err:     err,
	}
	b(req)
	return req.exec()
}

func Get(r *Request) {
	r.Method = http.MethodGet
}

func Options(r *Request) {
	r.Method = http.MethodOptions
}

func Post(r *Request) {
	r.Method = http.MethodPost
}

func Delete(r *Request) {
	r.Method = http.MethodDelete
}

func Put(r *Request) {
	r.Method = http.MethodPut
}

func Patch(r *Request) {
	r.Method = http.MethodPatch
}

func SetPath(p string) Builder {
	return func(r *Request) {
		r.URL = r.URL.ResolveReference(&url.URL{Path: p})
	}
}

func URL(u string) Builder {
	URL, err := url.Parse(u)
	return SetErr(err).With(func(r *Request) {
		r.URL = URL
	})
}

func Path(p ...string) Builder {
	return func(r *Request) {
		p = append([]string{r.URL.Path}, p...)
		r.URL = r.URL.ResolveReference(&url.URL{Path: path.Join(p...)})
	}
}

func JSON(v interface{}) Builder {
	b, err := json.Marshal(v)
	return MultiBuilder(
		SetErr(err),
		BytesBody(b),
		ContentType("application/json"),
	)
}

func ContentType(v string) Builder {
	return SetHeader("Content-Type", v)
}

func SetErr(err error) Builder {
	return func(r *Request) {
		if err != nil {
			r.err = err
		}
	}
}

func Form(v url.Values) Builder {
	return MultiBuilder(
		BytesBody([]byte(v.Encode())),
		ContentType("application/x-www-form-urlencoded"),
	)
}

func BytesBody(b []byte) Builder {
	return func(r *Request) {
		r.Body = ioutil.NopCloser(bytes.NewReader(b))
	}
}

func SetHeader(k, v string) Builder {
	return func(r *Request) {
		r.Header.Set(k, v)
	}
}

func SetQuery(k, v string) Builder {
	return func(r *Request) {
		q := r.Request.URL.Query()
		q.Set(k, v)
		r.Request.URL.RawQuery = q.Encode()
	}
}

func SetClient(cli client) Builder {
	return func(r *Request) {
		r.c = cli
	}
}

func Bearer(t string) Builder {
	return SetHeader("authorization", "Bearer "+t)
}

func IfNoErr(b Builder) Builder {
	return func(r *Request) {
		if r.err == nil {
			b(r)
		}
	}
}

//
//type Execable interface {
//	Exec() Response
//}
//
//type ExecableFn func() Response
//
//func (f ExecableFn) Exec() Response {
//	return f()
//}
//
//func PrintJSONResponse(r Response) Builder {
//	b, err := ioutil.ReadAll(r.Response.Body)
//	r.Body = ioutil.NopCloser(bytes.NewReader(b))
//	var jsR JSObj
//	return SetErr(err).With(
//		SetErr(r.JSON(&jsR)),
//		func(*Request) { fmt.Println(jsR) },
//	)
//}
//
//const isDebug = false
//
//func debug(f string, args ...interface{}) {
//	if isDebug {
//		fmt.Printf(f, args...)
//		fmt.Println()
//	}
//}
//
//func Must(err error) {
//	if err != nil {
//		panic(err)
//	}
//}
