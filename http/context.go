/*
 * Revision History:
 *     Initial: 2018/05/26        ShiChao
 */

package xhttp

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

	IsTLS() bool

	IsWebSocket() bool

	IP() string

	FormValue(name string) string

	FormParams() (url.Values, error)

	Cookie(name string) (*http.Cookie, error)

	SetCookie(cookie *http.Cookie)

	Cookies() []*http.Cookie

	Get(key string) interface{}

	Set(key string, val interface{})

	JSON(code int, i interface{}) error

	JSONBody(v interface{}) error

	Redirect(code int, url string) error

	Reset(r *http.Request, w http.ResponseWriter)
}

type context struct {
	req   *http.Request
	res   http.ResponseWriter
	store map[string]interface{}
}

func newContext() (ctx Context) {
	return &context{}
}

func (c *context) Request() *http.Request {
	return c.req
}

func (c *context) SetRequest(r *http.Request) {
	c.req = r
}

func (c *context) Response() http.ResponseWriter {
	return c.res
}

func (c *context) IsTLS() bool {
	return c.req.TLS != nil
}

func (c *context) IsWebSocket() bool {
	upgrade := c.req.Header.Get(HeaderUpgrade)
	return upgrade == "websocket" || upgrade == "Websocket"
}

func (c *context) IP() string {
	return c.req.RemoteAddr
}

func (c *context) FormValue(name string) string {
	return c.req.FormValue(name)
}

func (c *context) FormParams() (url.Values, error) {
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

func (c *context) Cookie(name string) (*http.Cookie, error) {
	return c.req.Cookie(name)
}

func (c *context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response(), cookie)
}

func (c *context) Cookies() []*http.Cookie {
	return c.req.Cookies()
}

func (c *context) Get(key string) interface{} {
	return c.store[key]
}

func (c *context) Set(key string, val interface{}) {
	if c.store == nil {
		c.store = make(map[string]interface{})
	}
	c.store[key] = val
}

func (c *context) JSON(code int, v interface{}) (err error) {
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

func (c *context) JSONBody(v interface{}) error {
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

func (c *context) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return errInvalidRedirectCode
	}
	c.res.Header().Set(HeaderLocation, url)
	c.res.WriteHeader(code)
	return nil
}

func (c *context) Reset(r *http.Request, w http.ResponseWriter) {
	c.req = r
	c.res = w
	c.store = nil
}
