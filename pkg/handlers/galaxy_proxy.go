package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/psvmcc/hub/pkg/misc"
	"github.com/psvmcc/hub/pkg/types"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func GalaxyProxyCollection(key string) echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg := c.Get("cfg").(types.ConfigFile)
		logger := c.Get("logger").(*zap.SugaredLogger)
		loggerNS := "galaxy_proxy_connection"
		namespace := c.Param("namespace")
		name := c.Param("name")
		url := fmt.Sprintf("%s/api/v3/collections/%s/%s/", cfg.Server.Galaxy[key].URL, namespace, name)
		dest := fmt.Sprintf("%s/galaxy/%s/index/%s/%s/index.json", cfg.Dir, key, namespace, name)

		headers := types.RequestHeaders{
			"UserAgent": "ansible-galaxy",
		}
		status, err := misc.DownloadFile(url, dest, headers)
		if err != nil {
			logger.Named(loggerNS).Errorf("[Downloading] %s", err)
			if _, err = os.Stat(dest); errors.Is(err, os.ErrNotExist) {
				logger.Named(loggerNS).Errorf("[FS]: %s", err)
				return c.String(status, fmt.Sprintf("%v", err))
			}
			c.Response().Header().Add("X-Cache-Status", "HIT")
			logger.Named(loggerNS).Debugf("Remote %s served from local file %s", url, dest)
		} else {
			logger.Named(loggerNS).Debugf("Remote %s saved as %s", url, dest)
		}
		var collection types.GalaxyCollection
		err = collection.ReadFromJSONFile(dest)
		if err != nil {
			logger.Named(loggerNS).Errorf("Unable to parse local json file %s, got error: %s", dest, err)
		}
		collection.Href = fmt.Sprintf("/galaxy/%s/api/v3/collections/%s/%s/", key, namespace, name)
		collection.VersionsURL = fmt.Sprintf("/galaxy/%s/api/v3/collections/%s/%s/versions/", key, namespace, name)
		collection.HighestVersion.Href = fmt.Sprintf("/galaxy/%s/api/v3/collections/%s/%s/versions/%s/", key, namespace, name, collection.HighestVersion.Version)
		return c.JSON(http.StatusOK, collection)
	}
}

func GalaxyProxyCollectionVersions(key string) echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg := c.Get("cfg").(types.ConfigFile)
		logger := c.Get("logger").(*zap.SugaredLogger)
		loggerNS := "galaxy_proxy_connection_versions"
		namespace := c.Param("namespace")
		name := c.Param("name")
		url := fmt.Sprintf("%s/api/v3/collections/%s/%s/versions/?%s", cfg.Server.Galaxy[key].URL, namespace, name, c.QueryString())
		dest := fmt.Sprintf("%s/galaxy/%s/index/%s/%s/versions/index/%s", cfg.Dir, key, namespace, name, c.QueryString())

		headers := types.RequestHeaders{
			"UserAgent": "ansible-galaxy",
		}
		status, err := misc.DownloadFile(url, dest, headers)
		if err != nil {
			logger.Named(loggerNS).Errorf("[Downloading] %s", err)
			if _, err = os.Stat(dest); errors.Is(err, os.ErrNotExist) {
				logger.Named(loggerNS).Errorf("[FS]: %s", err)
				return c.String(status, fmt.Sprintf("%v", err))
			}
			c.Response().Header().Add("X-Cache-Status", "HIT")
			logger.Named(loggerNS).Debugf("Remote %s served from local file %s", url, dest)
		} else {
			logger.Named(loggerNS).Debugf("Remote %s saved as %s", url, dest)
		}
		var collectionVersions types.GalaxyCollectionVersions
		err = collectionVersions.ReadFromJSONFile(dest, key, namespace, name)
		if err != nil {
			logger.Named(loggerNS).Errorf("Unable to parse local json file %s, got error: %s", dest, err)
		}
		return c.JSON(http.StatusOK, collectionVersions)
	}
}

func GalaxyProxyCollectionVersionInfo(key string) echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg := c.Get("cfg").(types.ConfigFile)
		logger := c.Get("logger").(*zap.SugaredLogger)
		loggerNS := "galaxy_proxy_connection_version_info"
		namespace := c.Param("namespace")
		name := c.Param("name")
		version := c.Param("version")
		url := fmt.Sprintf("%s/api/v3/collections/%s/%s/versions/%s", cfg.Server.Galaxy[key].URL, namespace, name, version)
		dest := fmt.Sprintf("%s/galaxy/%s/index/%s/%s/versions/%s/index.json", cfg.Dir, key, namespace, name, version)

		scheme := c.Scheme()
		host := c.Request().Host

		headers := types.RequestHeaders{
			"UserAgent": "ansible-galaxy",
		}
		_, err := misc.DownloadFile(url, dest, headers)
		if err != nil {
			logger.Named(loggerNS).Errorf("[Downloading] %s", err)
			if _, err = os.Stat(dest); errors.Is(err, os.ErrNotExist) {
				logger.Named(loggerNS).Errorf("[FS]: %s", err)
				return c.String(http.StatusNotFound, "")
			}
			c.Response().Header().Add("X-Cache-Status", "HIT")
			logger.Named(loggerNS).Debugf("Remote %s served from local file %s", url, dest)
		} else {
			logger.Named(loggerNS).Debugf("Remote %s saved as %s", url, dest)
		}
		var CollectionVersionInfo types.GalaxyCollectionVersionInfo
		err = CollectionVersionInfo.ReadFromJSONFile(dest)
		if err != nil {
			logger.Named(loggerNS).Errorf("Unable to parse local json file %s, got error: %s", dest, err)
		}
		CollectionVersionInfo.Href = fmt.Sprintf("/galaxy/%s/api/v3/collections/%s/%s/versions/%s/", key, namespace, name, version)
		CollectionVersionInfo.Collection.Href = fmt.Sprintf("/galaxy/%s/api/v3/collections/%s/%s/", key, namespace, name)
		CollectionVersionInfo.DownloadURL = fmt.Sprintf("%s://%s/galaxy/%s/get/%s/%s/%s", scheme, host, key, namespace, name, version)
		return c.JSON(http.StatusOK, CollectionVersionInfo)
	}
}

func GalaxyProxyCollectionGet(key string) echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg := c.Get("cfg").(types.ConfigFile)
		logger := c.Get("logger").(*zap.SugaredLogger)
		loggerNS := "galaxy_proxy_connection_get"
		namespace := c.Param("namespace")
		name := c.Param("name")
		version := strings.TrimRight(c.Param("version"), "/")
		versionFile := fmt.Sprintf("%s/galaxy/%s/index/%s/%s/versions/%s/index.json", cfg.Dir, key, namespace, name, version)
		var CollectionVersionInfo types.GalaxyCollectionVersionInfo
		err := CollectionVersionInfo.ReadFromJSONFile(versionFile)
		if err != nil {
			logger.Named(loggerNS).Debugf("Parse local json file %s, got error: %s", versionFile, err)

			url := fmt.Sprintf("%s/api/v3/collections/%s/%s/versions/%s", cfg.Server.Galaxy[key].URL, namespace, name, version)

			headers := types.RequestHeaders{
				"UserAgent": "ansible-galaxy",
			}
			_, err := misc.DownloadFile(url, versionFile, headers)
			if err != nil {
				logger.Named(loggerNS).Errorf("[Downloading] %s", err)
				return c.String(http.StatusBadRequest, "Downloading error")
			}
			logger.Named(loggerNS).Debugf("Remote %s saved as %s", url, versionFile)
			err = CollectionVersionInfo.ReadFromJSONFile(versionFile)
			if err != nil {
				logger.Named(loggerNS).Errorf("Unable to parse local json file %s, got error: %s", versionFile, err)
				return c.String(http.StatusBadRequest, "Metadata error")
			}
		}
		url := CollectionVersionInfo.DownloadURL
		dest := fmt.Sprintf("%s/galaxy/%s/binary/%s/%s/%s-%s-%s.tar.gz", cfg.Dir, key, namespace, name, namespace, name, version)

		headers := types.RequestHeaders{
			"UserAgent": "ansible-galaxy",
		}
		if _, err := os.Stat(dest); errors.Is(err, os.ErrNotExist) {
			status, err := misc.DownloadFile(url, dest, headers)
			if err != nil {
				logger.Named(loggerNS).Errorf("[Downloading] %s", err)
				return c.String(status, fmt.Sprintf("%v", err))
			}
			c.Response().Header().Add("X-Cache-Status", "MISS")
		} else {
			logger.Named(loggerNS).Debugf("Remote %s served from local file %s", url, dest)
			localSha, err := misc.CalculateSHA256(dest)
			if err != nil {
				logger.Named(loggerNS).Errorf("SHA calculating for %s error: %s", dest, err)
			}
			if CollectionVersionInfo.Artifact.Sha256 == localSha {
				c.Response().Header().Add("X-Cache-Status", "HIT")
			} else {
				logger.Named(loggerNS).Errorf("SHA mismatch for %s local %s and remote %s", dest, localSha, CollectionVersionInfo.Artifact.Sha256)
				status, err := misc.DownloadFile(url, dest, headers)
				if err != nil {
					logger.Named(loggerNS).Errorf("[Downloading] %s", err)
					return c.String(status, fmt.Sprintf("%v", err))
				}
				logger.Named(loggerNS).Debugf("Downloaded %s", url)
				c.Response().Header().Add("X-Cache-Status", "EXPIRED")
			}
		}

		c.Response().Header().Add("Content-Type", "application/gzip")
		c.Response().Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s-%s-%s.tar.gz\"", namespace, name, version))
		return c.File(dest)
	}
}
