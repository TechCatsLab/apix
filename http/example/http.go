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
	app := xhttp.New()
	app.Get("/", handle)

	app.Use(negroni.NewLogger())
	app.RunTLS(":3344", "../certs/server.crt", "../certs/server.key")
}

func handle(ctx xhttp.Context) error {
	//panic("panic test")

	res := ctx.Response()

	res.WriteHeader(200)
	res.Write([]byte("hello world! \n"))
	return nil
}
