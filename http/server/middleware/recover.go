/*
 * Revision History:
 *     Initial: 2018/05/27        Wang Riyu
 */

package middleware

import (
	"fmt"
	"net/http"

	"github.com/urfave/negroni"
)

// NegroniRecoverHandler returns a handler for recover from a http request.
func NegroniRecoverHandler() negroni.Handler {
	fn := func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		defer recoverFunc(w)
		next.ServeHTTP(w, r)
	}
	return negroni.HandlerFunc(fn)
}

func recoverFunc(w http.ResponseWriter) {
	if err := recover(); err != nil {
		fmt.Println("Recovered from panic in http handler:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
