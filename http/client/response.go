package client

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/TechCatsLab/apix/file"
)

// Response represents the response from an HTTP request.
type Response struct {
	*http.Response
}

func saveFile(resp *Response, target string) (written int64, err error) {
	defer resp.Body.Close()

	var name, ext, filename string

	target, err = filepath.Abs(target)
	if err != nil {
		return 0, err
	}

	name = file.FileName(resp.Request.URL.String())
	ext = path.Ext(resp.Request.URL.String())

	if ext == "" {
		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			exts, err := mime.ExtensionsByType(contentType)
			if err == nil {
				ext = exts[0]
			}
		}
	}

	filename = fmt.Sprintf("%s/%s%s", target, name, ext)
	for i := 1; file.IsExist(filename); i++ {
		filename = fmt.Sprintf("%s/%s(%d)%s", target, name, i, ext)
	}
	fmt.Println(target, name, ext)

	if !file.IsDirExist(target) {
		return 0, fmt.Errorf("target directory (%s) cann't be found", target)
	}

	if !file.IsPermissionDenied(target) {
		return 0, fmt.Errorf("target directory (%s): permission denied", target)
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return io.Copy(file, resp.Body)
}
