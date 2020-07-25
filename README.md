```go
package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/icimpoes/hcli"
)

var apiRequest = hcli.New(
	hcli.URL("http://my-service.com"),
	hcli.Path("api", "v1"),
	SetHeaders,
)

// custom request builder
func SetHeaders(req *hcli.Request) {
	req.Header.Set("H-1", "1")
	req.Header.Set("H-2", "2")
}

func login(username, password string) hcli.Builder {
	credentials := url.Values{}
	credentials.Set("username", username)
	credentials.Set("password", password)

	return apiRequest.With(
		hcli.Post,
		hcli.Path("/token"),
		hcli.Form(credentials))
}

func users(token string) hcli.Builder {
	return apiRequest.With(
		hcli.Get,
		hcli.Path("user"),
		hcli.Bearer(token))
}

func main() {
	resp := login("user", "pass").
		Chain(func(resp hcli.Response) hcli.Builder {
			if resp.StatusCode != http.StatusOK {
				return hcli.SetErr(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
			}
			var body hcli.JSObj
			return hcli.New(
				hcli.SetErr(resp.JSON(&body)),
				users(body.Str("access_token")))
		}).Exec()

	var body hcli.JSObj
	err := resp.JSON(&body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(body)
}

```
