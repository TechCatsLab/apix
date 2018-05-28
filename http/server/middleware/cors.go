/*
 * Revision History:
 *     Initial: 2018/05/28        ShiChao
 */

package middleware

import (
	"github.com/urfave/negroni"
	"github.com/rs/cors"
)

func NegroniCorsAllowAll() negroni.Handler {
	return cors.AllowAll()
}

func NegroniCorsNew(opt cors.Options) negroni.Handler {
	return cors.New(opt)
}