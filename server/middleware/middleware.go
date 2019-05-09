// provide customer middleware which common task
package middleware

import (
	"github.com/by46/whalefs/model"
	"github.com/labstack/echo"
)

type ExtendContext struct {
	echo.Context
	FileParams *model.FileParams
}
