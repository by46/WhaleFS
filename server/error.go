package server

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"

	"github.com/by46/whalefs/common"
)

func (s *Server) HTTPErrorHandler(err error, ctx echo.Context) {
	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)
	switch err.(type) {
	case *echo.HTTPError:
		he := err.(*echo.HTTPError)
		code = he.Code
		msg = he.Message
		if he.Internal != nil {
			err = fmt.Errorf("%v, %v", err, he.Internal)
		}
	case *common.BusinessError:
		e := err.(*common.BusinessError)
		switch e.Code {
		case common.CodeFileNotExists:
			code = http.StatusNotFound
		case common.CodeBucketNotExists:
			code = http.StatusForbidden
		case common.CodeForbidden:
			code = http.StatusForbidden
		default:
			code = http.StatusInternalServerError
		}
		msg = http.StatusText(code)
	default:
		if s.Debug {
			msg = err.Error()
		}
		msg = http.StatusText(code)
	}

	if _, ok := msg.(string); ok {
		msg = echo.Map{"message": msg}
	}

	// Send response
	if !ctx.Response().Committed {
		if ctx.Request().Method == http.MethodHead { // Issue #608
			err = ctx.NoContent(code)
		} else {
			err = ctx.JSON(code, msg)
		}
	}

	if err != nil {
		s.Logger.Error(err)
	}
}
