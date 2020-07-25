package hcli

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

var (
	apiRequestBuilder = New(
		URL("http://hello.com"),
		Path("api/v1"),
		SetHeader("x-h-1", "123"),
		SetHeader("x-h-2", "321"),
		SetClient(DumpClient{Client: http.DefaultClient}),
	)
)

func Test_PostUser(t *testing.T) {
	gock.New("http://hello.com").
		Post("/api/v1/user").
		MatchHeaders(map[string]string{
			"x-h-1": "123",
			"x-h-2": "321",
		}).
		JSON(JSObj{
			"name": "xxx",
			"age":  42,
		}).
		Reply(http.StatusNoContent)

	resp := apiRequestBuilder.With(Path("/user"), Post, JSON(JSObj{
		"name": "xxx",
		"age":  42,
	})).Exec()

	require.NoError(t, resp.Err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.False(t, gock.HasUnmatchedRequest())
	gock.Clean()
}

func Test_GetUsers(t *testing.T) {
	expected := []JSObj{
		{"id": float64(1), "name": "xxx"},
		{"id": float64(2), "name": "yyy"},
	}
	gock.New("http://hello.com").
		Get("/api/v1/user").
		MatchHeaders(map[string]string{
			"x-h-1": "123",
			"x-h-2": "321",
			"x-h-3": "321",
		}).
		Reply(http.StatusOK).
		JSON(expected)

	resp := apiRequestBuilder.
		With(Path("/user"), Get, SetHeader("x-h-3", "321")).
		Exec()

	require.NoError(t, resp.Err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var responseBody []JSObj
	assert.NoError(t, resp.JSON(&responseBody))
	assert.Equal(t, expected, responseBody)
	assert.False(t, gock.HasUnmatchedRequest())
	gock.Clean()
}

func Test_Chain(t *testing.T) {
	credentials := url.Values{}
	credentials.Set("username", "user")
	credentials.Set("password", "p@s")

	expected := []JSObj{
		{"id": float64(1), "name": "xxx"},
		{"id": float64(2), "name": "yyy"},
	}
	gock.New("http://hello.com").
		Post("/api/v1/token").
		MatchHeaders(map[string]string{
			"x-h-1": "123",
			"x-h-2": "321",
		}).
		BodyString(credentials.Encode()).
		Reply(http.StatusOK).
		JSON(JSObj{"access_token": "token_123_456"})
	gock.New("http://hello.com").
		Get("/api/v1/user").
		MatchHeaders(map[string]string{
			"x-h-1":         "123",
			"x-h-2":         "321",
			"Authorization": "Bearer token_123_456",
		}).
		Reply(http.StatusOK).
		JSON(expected)

	login := apiRequestBuilder.With(
		Path("/token"), Post, Form(credentials))

	resp := login.Chain(func(resp Response) Builder {
		if resp.StatusCode != http.StatusOK {
			return SetErr(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
		}
		var body JSObj
		return New(SetErr(resp.JSON(&body)),
			apiRequestBuilder.With(Get, Path("user"), Bearer(body.Str("access_token"))))
	}).Exec()

	require.NoError(t, resp.Err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var responseBody []JSObj
	assert.NoError(t, resp.JSON(&responseBody))
	assert.Equal(t, expected, responseBody)
	assert.False(t, gock.HasUnmatchedRequest())
	gock.Clean()
}

func Test_BuilderError(t *testing.T) {
	resp := New(URL("%//"), Get, SetHeader("a", "b")).Exec()

	assert.EqualError(t, resp.Err, "parse \"%//\": invalid URL escape \"%//\"")
	assert.False(t, gock.HasUnmatchedRequest())
	gock.Clean()
}

func Test_ChainError(t *testing.T) {
	resp := New(URL("%//"), Get).
		Chain(func(Response) Builder {
			return apiRequestBuilder.With(Get)
		}).Exec()

	assert.EqualError(t, resp.Err, "parse \"%//\": invalid URL escape \"%//\"")
	assert.False(t, gock.HasUnmatchedRequest())
	gock.Clean()
}
