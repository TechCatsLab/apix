/*
 * Revision History:
 *     Initial: 2018/05/26        ShiChao
 */

package main

import (
	"github.com/TechCatsLab/apix/http"
	"github.com/urfave/negroni"
)

func main() {
	server := http.New()
	server.Get("/", handle)

	server.Use(negroni.NewLogger())
	server.RunTLS(":3344", "./certs/server.crt", "./certs/server.key")
}

func handle(ctx http.Context) error {
	//panic("panic test")

	res := ctx.Response()

	res.WriteHeader(200)
	res.Write([]byte("hello world! \n"))
	return nil
}
