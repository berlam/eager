package pkg

import (
	"io"
	"net/http"
	"net/url"
)

func NewHttpClient() *http.Client {
	return &http.Client{}
}

func CreateJsonRequest(client *http.Client, httpMethod string, server *url.URL, userinfo *url.Userinfo, payload io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(httpMethod, server.String(), payload)
	if err != nil {
		return nil, err
	}
	password, _ := userinfo.Password()
	request.SetBasicAuth(userinfo.Username(), password)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	return response, err
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}
