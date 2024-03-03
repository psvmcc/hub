package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/psvmcc/hub/pkg/misc"
	"github.com/psvmcc/hub/pkg/types"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func PypiSimple(key string) echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg := c.Get("cfg").(types.ConfigFile)
		logger := c.Get("logger").(*zap.SugaredLogger)
		loggerNS := "pypi_simple"
		name := c.Param("name")
		url := fmt.Sprintf("%s/%s/", cfg.Server.PYPI[key], name)
		dest := fmt.Sprintf("%s/pypi/%s/%s/index.json", cfg.Dir, key, name)

		scheme := c.Scheme()
		host := c.Request().Host
		headers := types.RequestHeaders{
			"UserAgent": "pypi",
			"Accept":    "application/vnd.pypi.simple.v1+json",
		}

		status, err := misc.DownloadFile(url, dest, headers)
		if err != nil {
			logger.Named(loggerNS).Errorf("[Downloading] %s", err)
			if _, err = os.Stat(dest); errors.Is(err, os.ErrNotExist) {
				logger.Named(loggerNS).Errorf("[FS]: %s", err)
				return c.String(status, "Please check logs...")
			}
			c.Response().Header().Add("X-Cache-Status", "HIT")
			logger.Named(loggerNS).Debugf("Remote %s served from local file %s", url, dest)
		} else {
			logger.Named(loggerNS).Debugf("Remote %s saved as %s", url, dest)
		}

		c.Response().Header().Add("Content-Type", "text/html")
		var pypiMetadata types.PypiMetadata
		err = pypiMetadata.ReadFromJSONFile(dest)
		if err != nil {
			logger.Named(loggerNS).Errorf("Unable to parse local json file %s, got error: %s", dest, err)
		}

		for i := range pypiMetadata.Files {
			pypiMetadata.Files[i].URL = fmt.Sprintf("%s://%s/pypi/%s/packages/%s/%s", scheme, host, key, name, pypiMetadata.Files[i].Filename)
		}
		return c.Render(http.StatusOK, "pypi", pypiMetadata)
	}
}

func PypiPackages(key string) echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg := c.Get("cfg").(types.ConfigFile)
		logger := c.Get("logger").(*zap.SugaredLogger)
		loggerNS := "pypi_packages"
		name := c.Param("name")
		filename := c.Param("filename")
		indexDest := fmt.Sprintf("%s/pypi/%s/%s/index.json", cfg.Dir, key, name)

		dest := fmt.Sprintf("%s/pypi/%s/%s/%s", cfg.Dir, key, name, filename)
		var url, sha string

		headers := types.RequestHeaders{
			"UserAgent": "pypi",
		}

		var pypiMetadata types.PypiMetadata
		err := pypiMetadata.ReadFromJSONFile(indexDest)
		if err != nil {
			logger.Named(loggerNS).Debugf("Parse local json file %s, got error: %s", indexDest, err)

			url = fmt.Sprintf("%s/%s/", cfg.Server.PYPI[key], name)

			headers = types.RequestHeaders{
				"UserAgent": "pypi",
				"Accept":    "application/vnd.pypi.simple.v1+json",
			}
			_, err := misc.DownloadFile(url, indexDest, headers)
			if err != nil {
				logger.Named(loggerNS).Errorf("[Downloading] %s", err)
				return c.String(http.StatusBadRequest, "Downloading error")
			}
			logger.Named(loggerNS).Debugf("Remote %s saved as %s", url, indexDest)
			err = pypiMetadata.ReadFromJSONFile(indexDest)
			if err != nil {
				logger.Named(loggerNS).Errorf("Unable to parse local json file %s, got error: %s", indexDest, err)
				return c.String(http.StatusBadRequest, "Metadata error")
			}
		}

		for i := range pypiMetadata.Files {
			if pypiMetadata.Files[i].Filename == filename {
				url = pypiMetadata.Files[i].URL
				sha = pypiMetadata.Files[i].Hashes.Sha256
				break
			}
		}

		if url == "" {
			logger.Named(loggerNS).Errorf("URL is empty for %s/%s", name, filename)
			return c.String(http.StatusNotFound, fmt.Sprintf("URL is empty for %s/%s", name, filename))
		}

		if _, err := os.Stat(dest); errors.Is(err, os.ErrNotExist) {
			status, err := misc.DownloadFile(url, dest, headers)
			if err != nil {
				logger.Named(loggerNS).Errorf("Downloading %s error: %s", url, err)
				return c.String(status, fmt.Sprintf("%v", err))
			}
			logger.Named(loggerNS).Debugf("Local file %s not found", dest)
			c.Response().Header().Add("X-Cache-Status", "MISS")
		} else {
			localSha, err := misc.CalculateSHA256(dest)
			if err != nil {
				logger.Named(loggerNS).Errorf("SHA calculating for %s error: %s", dest, err)
			}
			if sha == localSha {
				c.Response().Header().Add("X-Cache-Status", "HIT")
			} else {
				logger.Named(loggerNS).Errorf("SHA mismatch for %s local %s and remote %s", dest, localSha, sha)
				status, err := misc.DownloadFile(url, dest, headers)
				if err != nil {
					logger.Named(loggerNS).Errorf("[Downloading] %s", err)
					return c.String(status, fmt.Sprintf("%v", err))
				}
				logger.Named(loggerNS).Debugf("Remote %s saved as %s", url, dest)
				c.Response().Header().Add("X-Cache-Status", "EXPIRED")
			}
		}

		c.Response().Header().Add("Content-Type", "application/gzip")
		c.Response().Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
		return c.File(dest)
	}
}
