package utils

import (
	"io/ioutil"
	"net/http"
	url2 "net/url"
)

const (
	HeaderETag         = "Etag"
	HeaderIfNoneMatch  = "If-None-Match"
	HeaderExpires      = "Expires"
	HeaderCacheControl = "Cache-Control"
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
	client := &http.Client{}

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
