/*
 * Revision History:
 *     Initial: 2018/5/27        ShiChao
 */

package main

import (
	"github.com/fengyfei/gu/libs/http/server"
	"github.com/fengyfei/gu/libs/http/server/middleware"
	"github.com/fengyfei/gu/libs/logger"
)

type echo struct {
	Message *string `json:"message"`
}

func echoHandler(c *server.Context) error {
	var (
		err error
		req echo
	)

	if err = c.JSONBody(&req); err != nil {
		return err
	}

	return c.ServeJSON(&req)
}

func postHandler(c *server.Context) error {
	// w.Write([]byte("Post\n"))
	return nil
}

func panicHandler(c *server.Context) error {
	panic("Panic testing")
	return nil
}

// curl command
// curl -k https://127.0.0.1:9574/
// curl -k -X POST https://127.0.0.1:9574/post
// curl -k https://127.0.0.1:9574/panic
func https() {
	configuration := &server.Configuration{
		Address: "0.0.0.0:9574",
	}

	tlsConfig := &server.TLSConfiguration{
		Key:  "../certs/server.key",
		Cert: "../certs/server.crt",
	}

	router := server.NewRouter()
	router.Get("/", echoHandler)
	router.Post("/post", postHandler)
	router.Get("/panic", panicHandler)

	ep := server.NewEntrypoint(configuration, tlsConfig)
	cors := middleware.CORSAllowAll()

	// add middlewares
	ep.AttachMiddleware(middleware.NegroniRecoverHandler())
	ep.AttachMiddleware(middleware.NegroniLoggerHandler())
	ep.AttachMiddleware(cors)

	if err := ep.Start(router.Handler()); err != nil {
		logger.Error(err)
		return
	}

	ep.Wait()
}

func main() {
	https()
}
