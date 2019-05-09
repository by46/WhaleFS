package server

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

func (s *Server) HTTPErrorHandler(err error, ctx echo.Context) {
	ctx.Response().WriteHeader(http.StatusInternalServerError)
	_, err1 := ctx.Response().Write([]byte(err.Error()))
	fmt.Printf("%s", err1)
}
