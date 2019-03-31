package bcs

import (
	"bytes"
	"eager/pkg"
	"fmt"
	"github.com/magiconair/properties/assert"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestGetBcsEffort(t *testing.T) {
	testUrl := &url.URL{
		Scheme: "http",
		Host:   "localhost",
	}
	testDate := time.Now()
	testReport := ""

	client := pkg.NewTestClient(func(req *http.Request) *http.Response {
		parse := func(url *url.URL, path string, args ...string) *url.URL {
			parse, _ := url.Parse(fmt.Sprintf(path, args))
			return parse
		}

		switch req.URL {
		case parse(testUrl, bcsLogin):
			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		case parse(testUrl, bcsLogout):
			assert.Equal(t, req.Header.Get("Authorization"), "")
			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		case parse(testUrl, bcsShowEffort):
			assert.Equal(t, req.Header.Get("Authorization"), "")
			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		case parse(testUrl, bcsGetEffort):
			assert.Equal(t, req.Header.Get("Authorization"), "")
			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
				// Must be set to non-nil value or it panics
				Header: make(http.Header),
			}
		default:
			return &http.Response{
				StatusCode: 404,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`Not found`)),
				Header:     make(http.Header),
			}
		}
	})

	timesheet := GetTimesheet(client, testUrl, nil, testDate.Year(), testDate.Month(), testReport)
	assert.Equal(t, len(timesheet), 1)
}
