package utils

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo"
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
	StatusCode int
	Header     http.Header
	body       io.ReadCloser
	Content    []byte
}

func (r *Response) Json(v interface{}) (err error) {
	if r.Content == nil {
		r.Content, err = ioutil.ReadAll(r.body)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return errors.WithStack(json.Unmarshal(r.Content, v))
}

func (r *Response) Error() error {
	if r.StatusCode >= http.StatusOK && r.StatusCode < http.StatusBadRequest {
		return nil
	}
	buf := make([]byte, ErrorResponseSize)
	n, _ := r.body.Read(buf)

	return errors.Errorf("status: %d, message: %s", r.StatusCode, string(buf[:n]))
}

func (r *Response) Read(p []byte) (n int, err error) {
	return r.body.Read(p)
}

func (r *Response) Close() error {
	return r.body.Close()
}

func (r *Response) ReadAll() ([]byte, error) {
	return ioutil.ReadAll(r)
}

func Get(url string, headers http.Header) (*Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if headers != nil {
		req.Header = headerCopy(req.Header, headers)
	}
	return do(req)
}

func Post(url string, headers http.Header, body io.Reader) (*Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if headers != nil {
		req.Header = headerCopy(req.Header, headers)
	}
	return do(req)
}

func do(req *http.Request) (*Response, error) {
	resp, err := client.Do(req)
	//if resp != nil {
	//	defer func() {
	//		_ = resp.Body.Close()
	//	}()
	//}
	if err != nil {
		if resp != nil {
			_ = resp.Body.Close()
		}
		return nil, errors.WithStack(err)
	}
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return nil, errors.WithStack(err)
	//}
	response := &Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		body:       resp.Body,
		//Content:    body,
	}
	return response, errors.WithStack(response.Error())
}

func headerCopy(dst, src http.Header) http.Header {
	for key := range src {
		value := src.Get(key)
		dst.Set(key, value)
	}
	return dst
}

func QueryExists(value url.Values, name string) (exists bool) {
	v := map[string][]string(value)
	_, exists = v[name]
	return
}

func Params(ctx echo.Context, name string) string {
	value := strings.TrimSpace(ctx.QueryParam(name))
	if value != "" {
		return value
	}
	return ctx.Request().FormValue(name)
}
