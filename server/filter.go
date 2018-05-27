/*
 * Revision History:
 *     Initial: 2018/5/26        ShiChao
 */

package server

import (
	"errors"
)

var (
	errFilterNotPassed = errors.New("filter check is not passed")
)

// FilterFunc defines a filter function which is invoked before the controller
// handler is executed.
type FilterFunc func(*Context) bool
