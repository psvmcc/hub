package misc

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/psvmcc/hub/pkg/types"
)

func DownloadFile(url, destination string, headers types.RequestHeaders) (code int, err error) {
	client := &http.Client{}

	var req *http.Request
	var response *http.Response

	req, err = http.NewRequest("GET", url, http.NoBody)
	if err != nil {
		code = http.StatusBadRequest
		return code, err
	}
	req.Header.Set("UserAgent", "hub")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	response, err = client.Do(req)
	if err != nil {
		code = http.StatusBadGateway
		return code, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		code = response.StatusCode
		return code, err
	}

	if err = os.MkdirAll(filepath.Dir(destination), 0o750); err != nil {
		err = fmt.Errorf("failed to create destination directory: %v", err)
		code = http.StatusConflict
		return code, err
	}

	n, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		code = http.StatusInternalServerError
		return code, err
	}

	tempFileName := fmt.Sprintf(".tmp.%s.%d.%d", filepath.Base(filepath.Clean(destination)), time.Now().UnixNano(), n.Int64())
	tempFilePath := fmt.Sprintf("%s/%s", filepath.Dir(filepath.Clean(destination)), tempFileName)
	tempFile, err := os.Create(filepath.Join(filepath.Dir(destination), filepath.Clean(tempFileName)))
	if err != nil {
		err = fmt.Errorf("failed to create temporary file: %v", err)
		code = http.StatusInternalServerError
		return code, err
	}
	defer os.Remove(tempFile.Name())

	_, err = io.Copy(tempFile, response.Body)
	if err != nil {
		err = fmt.Errorf("failed to copy response body to file: %v", err)
		code = http.StatusBadRequest
		return code, err
	}

	if lastModifiedHeader := response.Header.Get("Last-Modified"); lastModifiedHeader != "" {
		var lastModifiedTime time.Time
		if lastModifiedTime, err = time.Parse(http.TimeFormat, lastModifiedHeader); err == nil {
			var fileInfo os.FileInfo
			fileInfo, err = tempFile.Stat()
			if err != nil {
				err = fmt.Errorf("failed to get file info: %v", err)
				code = http.StatusInternalServerError
				return code, err
			}
			if err = os.Chtimes(tempFilePath, fileInfo.ModTime(), lastModifiedTime); err != nil {
				err = fmt.Errorf("failed to set last-modified time: %v", err)
				code = http.StatusInternalServerError
				return code, err
			}
		}
	}
	err = tempFile.Close()
	if err != nil {
		err = fmt.Errorf("tempFile close error: %v", err)
		code = http.StatusInternalServerError
		return code, err
	}

	if err := os.Rename(tempFile.Name(), destination); err != nil {
		err = fmt.Errorf("failed to rename temporary file to destination: %v", err)
		code = http.StatusInternalServerError
		return code, err
	}

	code = http.StatusOK
	return code, nil
}
