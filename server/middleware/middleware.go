// provide customer middleware which common task
package middleware

import (
	"github.com/labstack/echo"

	"github.com/by46/whalefs/model"
)

type ExtendContext struct {
	echo.Context
	FileParams *model.FileParams
}
