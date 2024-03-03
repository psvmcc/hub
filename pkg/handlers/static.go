package handlers

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/psvmcc/hub/pkg/misc"
	"github.com/psvmcc/hub/pkg/types"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func Static(key string) echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg := c.Get("cfg").(types.ConfigFile)
		logger := c.Get("logger").(*zap.SugaredLogger)
		loggerNS := "static"
		path := strings.TrimPrefix(c.Request().URL.String(), fmt.Sprintf("/static/%s/get/", key))
		url := fmt.Sprintf("%s/%s", cfg.Server.Static[key], path)
		dest := fmt.Sprintf("%s/static/%s/%s", cfg.Dir, key, path)

		headers := types.RequestHeaders{
			"UserAgent": "curl",
		}

		if _, err := os.Stat(dest); errors.Is(err, os.ErrNotExist) {
			c.Response().Header().Add("X-Cache-Status", "MISS")
		} else {
			equal, err := misc.FilesEqual(url, dest)
			if err != nil {
				logger.Named(loggerNS).Errorf("[FilesEqual]: %s", err)
			}

			if equal {
				c.Response().Header().Add("X-Cache-Status", "HIT")
				return c.File(dest)
			}
			c.Response().Header().Add("X-Cache-Status", "EXPIRE")
		}
		status, err := misc.DownloadFile(url, dest, headers)
		if err != nil {
			logger.Named(loggerNS).Errorf("[Downloading] %s", err)
			if _, err := os.Stat(dest); errors.Is(err, os.ErrNotExist) {
				logger.Named(loggerNS).Errorf("[FS]: %s", err)
				return c.String(status, "Please check logs...")
			}
			logger.Named(loggerNS).Debugf("Remote %s served from local file %s", url, dest)
		} else {
			logger.Named(loggerNS).Debugf("Remote %s saved as %s", url, dest)
		}
		return c.File(dest)
	}
}
