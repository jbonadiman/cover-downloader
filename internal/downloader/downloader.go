package downloader

import (
	"errors"
	"io"
	"net/http"
)

var ErrMissingFile = errors.New("missing file")

func DownloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrMissingFile
		}
		return nil, errors.New(resp.Status)
	}

	fileContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return fileContent, nil
}
