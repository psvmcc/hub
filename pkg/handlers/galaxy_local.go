package handlers

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/psvmcc/hub/pkg/misc"
	"github.com/psvmcc/hub/pkg/types"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func GalaxyLocalCollection(key string) echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg := c.Get("cfg").(types.ConfigFile)
		logger := c.Get("logger").(*zap.SugaredLogger)
		loggerNS := "galaxy_local_connection"
		namespace := c.Param("namespace")
		name := c.Param("name")

		dest := fmt.Sprintf("%s/%s/%s/", cfg.Server.Galaxy[key].Dir, namespace, name)
		if _, err := os.Stat(dest); errors.Is(err, os.ErrNotExist) {
			logger.Named(loggerNS).Debugf("Collection not found: %s/%s", namespace, name)
			return c.String(http.StatusNotFound, "No Collection found")
		}
		var collectionLocal types.GalaxyLocal
		err := collectionLocal.List(dest, namespace, name)
		if err != nil {
			logger.Named(loggerNS).Errorf("Collection list error: %s", err)
		}

		var collection types.GalaxyCollection
		collection.Href = fmt.Sprintf("/api/v3/collections/%s/%s/", namespace, name)
		collection.Namespace = namespace
		collection.Name = name
		collection.VersionsURL = fmt.Sprintf("/api/v3/collections/%s/%s/versions/", namespace, name)
		collection.HighestVersion.Version = collectionLocal.Latest.Version
		collection.HighestVersion.Href = fmt.Sprintf("/api/v3/collections/%s/%s/versions/%s/", namespace, name, collectionLocal.Latest.Version)
		collection.UpdatedAt = collectionLocal.Latest.Time.UTC()

		return c.JSON(http.StatusOK, collection)
	}
}

func GalaxyLocalCollectionVersions(key string) echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg := c.Get("cfg").(types.ConfigFile)
		logger := c.Get("logger").(*zap.SugaredLogger)
		loggerNS := "galaxy_local_versions"
		namespace := c.Param("namespace")
		name := c.Param("name")

		dest := fmt.Sprintf("%s/%s/%s/", cfg.Server.Galaxy[key].Dir, namespace, name)
		if _, err := os.Stat(dest); errors.Is(err, os.ErrNotExist) {
			logger.Named(loggerNS).Debugf("Collection not found: %s/%s", namespace, name)
			return c.String(http.StatusNotFound, "No Collection found")
		}
		var collectionLocal types.GalaxyLocal
		err := collectionLocal.List(dest, namespace, name)
		if err != nil {
			logger.Named(loggerNS).Errorf("Collection list error: %s", err)
		}

		var collectionVersions types.GalaxyCollectionVersions
		collectionVersions.Meta.Count = len(collectionLocal.Versions)
		for _, v := range collectionLocal.Versions {
			var verInfo types.GalaxyCollectionVersion
			verInfo.Version = v.Version
			verInfo.UpdatedAt = v.Time.UTC()
			verInfo.Href = fmt.Sprintf("/galaxy/%s/api/v3/collections/%s/%s/versions/%s/", key, namespace, name, v.Version)
			collectionVersions.Data = append(collectionVersions.Data, verInfo)
		}

		return c.JSON(http.StatusOK, collectionVersions)
	}
}

func GalaxyLocalCollectionVersionInfo(key string) echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg := c.Get("cfg").(types.ConfigFile)
		logger := c.Get("logger").(*zap.SugaredLogger)
		loggerNS := "galaxy_local_version_info"
		namespace := c.Param("namespace")
		name := c.Param("name")
		version := c.Param("version")

		scheme := c.Scheme()
		host := c.Request().Host

		dest := fmt.Sprintf("%s/%s/%s/", cfg.Server.Galaxy[key].Dir, namespace, name)
		if _, err := os.Stat(dest); errors.Is(err, os.ErrNotExist) {
			logger.Named(loggerNS).Debugf("Collection not found: %s/%s", namespace, name)
			return c.String(http.StatusNotFound, "No Collection found")
		}
		var collectionLocal types.GalaxyLocal
		err := collectionLocal.List(dest, namespace, name)
		if err != nil {
			logger.Named(loggerNS).Errorf("Collection list error: %s", err)
		}

		for _, v := range collectionLocal.Versions {
			if v.Version != version {
				continue
			} else if v.Version == version {

				return func(c echo.Context) error {
					file, err := os.Open(fmt.Sprintf("%s%s", dest, v.Filename))
					if err != nil {
						logger.Named(loggerNS).Errorf("Open tar error: %s", err)
					}
					defer file.Close()
					gzipReader, err := gzip.NewReader(file)
					if err != nil {
						logger.Named(loggerNS).Errorf("Reader gzip error: %s", err)
					}
					defer gzipReader.Close()
					tarReader := tar.NewReader(gzipReader)

					foundManifest := false
					foundFiles := false
					var manifest types.GalaxyCollectionVersionInfoManifest
					var files types.GalaxyCollectionVersionInfoFiles

					for {
						var header *tar.Header
						header, err = tarReader.Next()
						if err == io.EOF {
							break
						}
						if err != nil {
							logger.Named(loggerNS).Errorf("Error reading tar: %s", err)
						}
						if header.Name == "MANIFEST.json" {
							foundManifest = true
							decoder := json.NewDecoder(tarReader)
							err = decoder.Decode(&manifest)
							if err != nil {
								logger.Named(loggerNS).Errorf("Error parsing MANIFEST.json: %s", err)
							}
						}
						if header.Name == "FILES.json" {
							foundFiles = true
							decoder := json.NewDecoder(tarReader)
							err = decoder.Decode(&files)
							if err != nil {
								logger.Named(loggerNS).Errorf("Error parsing MANIFEST.json: %s", err)
							}
						}
						if foundManifest && foundFiles {
							break
						}
					}

					collectionVersionInfo := types.GalaxyCollectionVersionInfo{Signatures: []string{}}
					collectionVersionInfo.Version = version
					collectionVersionInfo.Href = fmt.Sprintf("/galaxy/%s/api/v3/collections/%s/%s/versions/%s/", key, namespace, name, version)
					collectionVersionInfo.UpdatedAt = v.Time.UTC()
					collectionVersionInfo.Name = name
					collectionVersionInfo.Namespace.Name = namespace
					collectionVersionInfo.Collection.Name = name
					collectionVersionInfo.Collection.Href = fmt.Sprintf("/galaxy/%s/api/v3/collections/%s/%s/", key, namespace, name)
					collectionVersionInfo.Artifact.Size = v.Size
					collectionVersionInfo.Artifact.Filename = v.Filename
					collectionVersionInfo.Artifact.Sha256, err = misc.CalculateSHA256(fmt.Sprintf("%s%s", dest, v.Filename))
					if err != nil {
						logger.Named(loggerNS).Errorf("sha calculating error for %s%s: %v", dest, v.Filename, err)
					}
					collectionVersionInfo.DownloadURL = fmt.Sprintf("%s://%s/galaxy/%s/get/%s/%s/%s", scheme, host, key, namespace, name, version)
					collectionVersionInfo.Manifest = manifest
					collectionVersionInfo.Metadata.Dependencies = manifest.CollectionInfo.Dependencies
					collectionVersionInfo.Files = files

					return c.JSON(http.StatusOK, collectionVersionInfo)
				}(c)
			}
			return c.String(http.StatusNotFound, "Not found")
		}

		return c.String(http.StatusNotFound, "Nothing found")
	}
}

func GalaxyLocalCollectionGet(key string) echo.HandlerFunc {
	return func(c echo.Context) error {
		cfg := c.Get("cfg").(types.ConfigFile)
		logger := c.Get("logger").(*zap.SugaredLogger)
		loggerNS := "galaxy_local_get"
		namespace := c.Param("namespace")
		name := c.Param("name")
		version := c.Param("version")

		dest := fmt.Sprintf("%s/%s/%s/%s-%s-%s.tar.gz", cfg.Server.Galaxy[key].Dir, namespace, name, namespace, name, version)
		if _, err := os.Stat(dest); errors.Is(err, os.ErrNotExist) {
			logger.Named(loggerNS).Debugf("Collection not found: %s/%s", namespace, name)
			return c.String(http.StatusNotFound, "")
		}
		c.Response().Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s-%s-%s.tar.gz\"", namespace, name, version))
		return c.File(dest)
	}
}
