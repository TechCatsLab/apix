/*
 * Revision History:
 *     Initial: 2018/05/29        Li Zebang
 */

package file

import (
	"errors"
	"os"
	"path"
	"strings"
)

var (
	// ErrUnwritable -
	ErrUnwritable = errors.New("file is not writable")
)

// IsExist -
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// IsDirExist -
func IsDirExist(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return info.IsDir()
}

// IsPermissionDenied -
func IsPermissionDenied(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsPermission(err)
	}
	return true
}

// FileName -
func FileName(name string) string {
	return strings.TrimSuffix(path.Base(name), path.Ext(name))
}
