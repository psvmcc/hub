package misc

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func FilesEqual(url, destination string) (bool, error) {
	localFileInfo, err := os.Stat(destination)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return nil
		},
	}

	response, err := client.Head(url)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HTTP request failed with status code: %d", response.StatusCode)
	}

	lastModifiedHeader := response.Header.Get("Last-Modified")
	remoteModTime, err := time.Parse(http.TimeFormat, lastModifiedHeader)
	if err != nil {
		return false, err
	}

	return localFileInfo.ModTime().Equal(remoteModTime) && localFileInfo.Size() == response.ContentLength, nil
}
