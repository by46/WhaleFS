package utils

import (
	"io/ioutil"
	"net/http"
	url2 "net/url"
	"time"
)

const (
	HeaderETag         = "Etag"
	HeaderIfNoneMatch  = "If-None-Match"
	HeaderExpires      = "Expires"
	HeaderCacheControl = "Cache-Control"
)

var (
	client = &http.Client{
		Timeout: 30 * time.Second,
	}
)

func Get(url string) ([]byte, http.Header, error) {
	u, err := url2.Parse(url)
	if err != nil {
		return nil, nil, err
	}
	req := &http.Request{
		Method: "GET",
		URL:    u,
	}

	resp, err := client.Do(req)
	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}
	if err != nil {
		return nil, nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	return body, resp.Header, err
}
