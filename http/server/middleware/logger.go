/*
 * Revision History:
 *     Initial: 2018/05/27        Wang Riyu
 */

package middleware

import (
	"github.com/urfave/negroni"
)

// NegroniLoggerHandler returns a logging handler.
func NegroniLoggerHandler() negroni.Handler {
	return negroni.NewLogger()
}
