package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	HeaderETag         = "Etag"
	HeaderIfNoneMatch  = "If-None-Match"
	HeaderExpires      = "Expires"
	HeaderCacheControl = "Cache-Control"
	ErrorResponseSize  = 512
)

var (
	client = &http.Client{
		Timeout: 30 * time.Second,
	}
)

type Response struct {
	*http.Response
	Content []byte
}

func (r *Response) Json(v interface{}) error {
	return json.Unmarshal(r.Content, v)
}

func (r *Response) Error() error {
	if r.StatusCode >= http.StatusOK && r.StatusCode < http.StatusBadRequest {
		return nil
	}

	return fmt.Errorf("status: %d, message: %s", r.StatusCode, string(r.Content[:ErrorResponseSize]))
}

func Get(url string, headers http.Header) (*Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if headers != nil {
		req.Header = HeaderCopy(req.Header, headers)
	}
	return do(req)
}

func Post(url string, headers http.Header, body io.Reader) (*Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if headers != nil {
		req.Header = HeaderCopy(req.Header, headers)
	}
	return do(req)
}

func do(req *http.Request) (*Response, error) {
	resp, err := client.Do(req)
	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	response := &Response{
		Response: resp,
		Content:  body,
	}
	if err := response.Error(); err != nil {
		return nil, errors.WithStack(err)
	}
	return response, nil
}

func HeaderCopy(dst, src http.Header) http.Header {
	for key := range src {
		value := src.Get(key)
		dst.Set(key, value)
	}
	return dst
}
