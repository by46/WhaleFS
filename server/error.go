package server

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

func (s *Server) HTTPErrorHandler(err error, ctx echo.Context) {
	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)

	realError := errors.Cause(err)
	switch realError.(type) {
	case *echo.HTTPError:
		he := realError.(*echo.HTTPError)
		code = he.Code
		msg = he.Message
		if he.Internal != nil {
			err = fmt.Errorf("%v, %v", err, he.Internal)
		}
	default:
		s.Logger.Errorf("%+v", err)
		if s.Debug {
			msg = err.Error()
		}
		msg = http.StatusText(code)
	}

	if _, ok := msg.(string); ok {
		msg = echo.Map{"message": msg, "state": "TYPE"}
	}

	// Send response
	if !ctx.Response().Committed {
		if ctx.Request().Method == http.MethodHead { // Issue #608
			err = ctx.NoContent(code)
		} else {
			err = ctx.JSON(code, msg)
		}

		if err != nil {
			s.Logger.Error(err)
		}
	}
}

func (s *Server) error(code int, err error) error {
	s.Logger.Error(err)
	return &echo.HTTPError{
		Code:    code,
		Message: err.Error(),
	}
}

func (s *Server) fatal(err error) error {
	return s.error(http.StatusInternalServerError, err)
}
