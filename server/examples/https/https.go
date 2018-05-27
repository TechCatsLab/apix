/*
 * MIT License
 *
 * Copyright (c) 2017 SmartestEE Co., Ltd.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

/*
 * Revision History:
 *     Initial: 2017/12/19        Feng Yifei
 *     Modify: 2017/12/20         Jia Chenhui
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
