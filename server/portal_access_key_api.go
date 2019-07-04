package server

import (
	"github.com/labstack/echo"
)

func (s *Server) createAccessKey(ctx echo.Context) (err error) {
	u := s.getCurrentUser(ctx)
	return err
}
