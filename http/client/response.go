/*
 * Revision History:
 *     Initial: 2018/05/28        Li Zebang
 */

package client

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/TechCatsLab/apix/file"
)

const (
	HeaderContentType = "Content-Type"

	MIMEApplicationJSON = "application/json"
	MIMEApplicationXML  = "application/xml"
	MIMETextXML         = "text/xml"
	MIMEApplicationForm = "application/x-www-form-urlencoded"
	MIMEMultipartForm   = "multipart/form-data"
)

var (
	ErrUnsupportedMediaType = errors.New("unsupported media type")
)

// Response represents the response from an HTTP request.
type Response struct {
	HTTPResponse *http.Response
}

// ToBytes -
func (r *Response) ToBytes() ([]byte, error) {
	defer r.HTTPResponse.Body.Close()
	return ioutil.ReadAll(r.HTTPResponse.Body)
}

// ToString -
func (r *Response) ToString() (string, error) {
	defer r.HTTPResponse.Body.Close()
	bs, err := r.ToBytes()
	return string(bs), err
}

// ToObject -
func (r *Response) ToObject(i interface{}) error {
	defer r.HTTPResponse.Body.Close()

	ctype := r.HTTPResponse.Header.Get(HeaderContentType)
	switch {
	case strings.HasPrefix(ctype, MIMEApplicationJSON):
		err := json.NewDecoder(r.HTTPResponse.Body).Decode(i)
		if err != nil {
			return err
		}
	case strings.HasPrefix(ctype, MIMEApplicationXML), strings.HasPrefix(ctype, MIMETextXML):
		err := xml.NewDecoder(r.HTTPResponse.Body).Decode(i)
		if err != nil {
			return err
		}
	default:
		return ErrUnsupportedMediaType
	}

	return nil
}

// SaveAsFile -
func (r *Response) SaveAsFile(directory string) (written int64, err error) {
	defer r.HTTPResponse.Body.Close()

	var name, ext, filename string

	directory, err = filepath.Abs(directory)
	if err != nil {
		return 0, err
	}

	name = file.FileName(r.HTTPResponse.Request.URL.String())
	ext = path.Ext(r.HTTPResponse.Request.URL.String())

	if ext == "" {
		contentType := r.HTTPResponse.Header.Get(HeaderContentType)
		if contentType == "" {
			exts, err := mime.ExtensionsByType(contentType)
			if err == nil {
				ext = exts[0]
			}
		}
	}

	filename = fmt.Sprintf("%s/%s%s", directory, name, ext)
	for i := 1; file.IsExist(filename); i++ {
		filename = fmt.Sprintf("%s/%s(%d)%s", directory, name, i, ext)
	}
	fmt.Println(directory, name, ext)

	if !file.IsDirExist(directory) {
		return 0, fmt.Errorf("target directory (%s) cann't be found", directory)
	}

	if !file.IsPermissionDenied(directory) {
		return 0, fmt.Errorf("target directory (%s): permission denied", directory)
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return io.Copy(file, r.HTTPResponse.Body)
}
