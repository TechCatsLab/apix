/*
 * Revision History:
 *     Initial: 2018/05/27        ShiChao
 */

package middleware

import (
	"net/http"
	"github.com/auth0/go-jwt-middleware"
	"github.com/urfave/negroni"
	"github.com/dgrijalva/jwt-go"
)

type Skipper func(path string) bool

// JWTMiddleware is a wrapper of go-jwt-middleware, but added a skipper func on it.
type JWTMiddleware struct {
	*jwtmiddleware.JWTMiddleware
	skipper Skipper
}

// handler runs HandlerWithNext func after skipper
func (jm *JWTMiddleware) handler(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	path := r.URL.Path
	if skip := jm.skipper(path); skip {
		next(w, r)
		return
	}

	jm.HandlerWithNext(w, r, next)
}

// NegroniJwtHandler returns a JWT middleware as a negroni handler. errHandler will be invoked if a err occurs while check JWT.
// and the errHandler must write to the response or not the client will be block.
func NegroniJwtHandler(key string, skipper Skipper, signMethod *jwt.SigningMethodHMAC, errHandler func(w http.ResponseWriter, r *http.Request, err string)) negroni.Handler {
	if signMethod == nil {
		signMethod = jwt.SigningMethodHS256
	}
	jm := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(key), nil
		},
		SigningMethod: signMethod,
		ErrorHandler:  errHandler,
	})

	if skipper == nil {
		skipper = defaulSkiper
	}

	JM := JWTMiddleware{
		jm,
		skipper,
	}

	return negroni.HandlerFunc(JM.handler)
}

func defaulSkiper(_ string) bool {
	return false
}
