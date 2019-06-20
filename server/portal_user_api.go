package server

import (
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

func (s *Server) addUser(ctx echo.Context) error {
	u, err := s.GetUser(ctx)
	if err != nil {
		return err
	}
	if u.Role != roleAdmin {
		return echo.NewHTTPError(http.StatusForbidden, "只有admin用户才能操作")
	}

	return nil
}

func (s *Server) listUser(ctx echo.Context) error {
	u, err := s.GetUser(ctx)
	if err != nil {
		return err
	}
	if u.Role != roleAdmin {
		return echo.NewHTTPError(http.StatusForbidden, "只有admin用户才能操作")
	}

	results, err := s.BucketMeta.GetAllUsers()

	if err != nil {
		s.Logger.Errorf("请求couchbase api失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	var infos []*basisInfo
	for {
		info := new(basisInfo)
		if results.Next(info) == false {
			break
		}
		info.Version = strconv.FormatUint(info.Cas, 10)
		infos = append(infos, info)
	}

	err = ctx.JSON(http.StatusOK, infos)
	return err
}

func (s *Server) updateUser(ctx echo.Context) error {
	u, err := s.GetUser(ctx)
	if err != nil {
		return err
	}
	if u.Role != roleAdmin {
		return echo.NewHTTPError(http.StatusForbidden, "只有admin用户才能操作")
	}

	return nil
}

func (s *Server) deleteUser(ctx echo.Context) error {
	u, err := s.GetUser(ctx)
	if err != nil {
		return err
	}
	if u.Role != roleAdmin {
		return echo.NewHTTPError(http.StatusForbidden, "只有admin用户才能操作")
	}

	return nil
}
