/*
 * Revision History:
 *     Initial: 2018/05/27        ShiChao
 */

package main

import (
	"fmt"
	"net/http"
	"github.com/TechCatsLab/apix/http/server"
	"github.com/TechCatsLab/apix/http/server/middleware"
	"github.com/dgrijalva/jwt-go"
)

const privateTokenKey = "your_token_key"

func main() {
	config := &server.Configuration{":3355"}
	ep := server.NewEntrypoint(config, nil)

	ep.AttachMiddleware(middleware.NegroniRecoverHandler())
	ep.AttachMiddleware(middleware.NegroniLoggerHandler())
	ep.AttachMiddleware(middleware.NegroniCorsAllowAll())
	ep.AttachMiddleware(middleware.NegroniJwtHandler(privateTokenKey, skipper, nil, jwtErrHandler))

	router := server.NewRouter()
	router.Get("/", handle)
	router.Post("/test", handle)

	ep.Start(router.Handler())

	ep.Run()
}

func handle(ctx *server.Context) error {
	//panic("panic test")

	res := ctx.Response()
	res.WriteHeader(200)
	res.Write([]byte("hello world! \n"))

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = &jwt.StandardClaims{}
	// private key generated with http://kjur.github.io/jsjws/tool_jwt.html
	s, e := token.SignedString([]byte(privateTokenKey))
	if e != nil {
		fmt.Println(e)
	}

	res.Write([]byte(s))

	return nil
}

func skipper(path string) bool {
	if path == "/skipper" {
		return true
	}
	return false
}

func jwtErrHandler(w http.ResponseWriter, r *http.Request, err string) {
	fmt.Println(err)
	http.Error(w, err, 401)
}
