/*
 * Revision History:
 *     Initial: 2018/05/26        ShiChao
 */

package http

import (
	"errors"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

var (
	errNoBody              = errors.New("request body is empty")
	errEmptyResponse       = errors.New("empty JSON response body")
	errNotJSONBody         = errors.New("request body is not JSON")
	errInvalidRedirectCode = errors.New("invalid redirect status code")
)

const (
	defaultMemory = 32 << 20 // 32 MB
)

type Context interface {
	Request() *http.Request

	Response() http.ResponseWriter

	FormValue(name string) string

	FormParams() (url.Values, error)

	Get(key string) interface{}

	Set(key string, val interface{})

	ServeJSON(code int, i interface{}) error

	ParseJSONBody(v interface{}) error
	
	Reset(r *http.Request, w http.ResponseWriter)
}

type ctx struct {
	req   *http.Request
	res   http.ResponseWriter
	store map[string]interface{}
}

func newContext() Context {
	return &ctx{}
}

func (c *ctx) Request() *http.Request {
	return c.req
}

func (c *ctx) SetRequest(r *http.Request) {
	c.req = r
}

func (c *ctx) Response() http.ResponseWriter {
	return c.res
}

func (c *ctx) FormValue(name string) string {
	return c.req.FormValue(name)
}

func (c *ctx) FormParams() (url.Values, error) {
	if strings.HasPrefix(c.req.Header.Get(HeaderContentType), MIMEMultipartForm) {
		if err := c.req.ParseMultipartForm(defaultMemory); err != nil {
			return nil, err
		}
	} else {
		if err := c.req.ParseForm(); err != nil {
			return nil, err
		}
	}
	return c.req.Form, nil
}

func (c *ctx) Get(key string) interface{} {
	return c.store[key]
}

func (c *ctx) Set(key string, val interface{}) {
	if c.store == nil {
		c.store = make(map[string]interface{})
	}
	c.store[key] = val
}

func (c *ctx) ServeJSON(code int, v interface{}) (err error) {
	if v == nil {
		return errEmptyResponse
	}

	resp, err := json.Marshal(v)
	if err != nil {
		return err
	}

	c.res.Header().Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	c.res.WriteHeader(code)
	_, err = c.res.Write(resp)
	return err
}

func (c *ctx) ParseJSONBody(v interface{}) error {
	if c.req.ContentLength == 0 {
		return errNoBody
	}

	if conType := c.req.Header.Get(HeaderContentType); !isJson(conType) {
		return errNotJSONBody
	}

	return json.NewDecoder(c.req.Body).Decode(v)
}

func isJson(s string) bool {
	s = strings.Replace(strings.ToUpper(s), " ", "", -1)
	j := strings.Replace(strings.ToUpper(MIMEApplicationJSON), " ", "", -1)
	jsonCharset := strings.Replace(strings.ToUpper(MIMEApplicationJSONCharsetUTF8), " ", "", -1)

	return strings.EqualFold(s, j) || strings.EqualFold(s, jsonCharset)
}

func (c *ctx) Reset(r *http.Request, w http.ResponseWriter) {
	c.req = r
	c.res = w
	c.store = nil
}
