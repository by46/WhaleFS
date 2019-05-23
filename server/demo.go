package server

import (
	"github.com/labstack/echo"
)

type Demo struct {
	Name string `json:"name"`
	Age  int32  `json:"age"`
}

func (s *Server) demo(ctx echo.Context) error {
	d := new(Demo)
	d.Age = 21
	err := ctx.Bind(d)
	return err
}
