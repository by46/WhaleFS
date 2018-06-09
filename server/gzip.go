package server

import (
	"github.com/labstack/echo"
	"strings"
	"compress/gzip"
	"io"
	"bytes"
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
		return err
	}
	_, err = io.Copy(rw, reader)
	if err != nil {
		return err
	}
	err = rw.Close()
	if err != nil {
		return nil
	}
	response.Header().Set(echo.HeaderContentEncoding, GzipScheme)
	response.Header().Add(echo.HeaderVary, echo.HeaderContentEncoding)
	response.Header().Set(echo.HeaderContentLength, string(buff.Len()))
	_, err = io.Copy(response.Writer, buff)
	return err
}