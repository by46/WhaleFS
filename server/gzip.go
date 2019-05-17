package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

func (s *Server) shouldGzip(ctx echo.Context) bool {
	header := ctx.Request().Header.Get(echo.HeaderAcceptEncoding)
	return strings.Contains(header, GzipScheme)
}

func (s *Server) compress(ctx echo.Context, reader io.Reader) error {
	response := ctx.Response()

	buff := bytes.NewBuffer(nil)
	rw, err := gzip.NewWriterLevel(buff, gzip.BestCompression)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = io.Copy(rw, reader)
	if err != nil {
		return errors.WithStack(err)
	}
	err = rw.Close()
	if err != nil {
		return errors.WithStack(err)
	}
	response.Header().Set(echo.HeaderContentEncoding, GzipScheme)
	response.Header().Add(echo.HeaderVary, echo.HeaderContentEncoding)
	response.Header().Set(echo.HeaderContentLength, string(buff.Len()))
	if _, err = io.Copy(response.Writer, buff); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
