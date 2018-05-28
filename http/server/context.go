/*
 * Revision History:
 *     Initial: 2018/5/27        ShiChao
 */

package server

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	json "github.com/json-iterator/go"
	"gopkg.in/go-playground/validator.v9"
)

const defaultMemory = 32 << 20 // 32 MB

var (
	errNoBody              = errors.New("request body is empty")
	errEmptyResponse       = errors.New("empty JSON response body")
	errNotJSONBody         = errors.New("request body is not JSON")
	errInvalidRedirectCode = errors.New("invalid redirect status code")
)

// Context wraps http.Request and http.ResponseWriter.
// Returned set to true if response was written to then don't execute other handler.
type Context struct {
	responseWriter http.ResponseWriter
	request        *http.Request
	Validator      *validator.Validate
	LastError      error
	store          map[string]interface{}
}

// NewContext create a new context.
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		responseWriter: w,
		request:        r,
		store:          make(map[string]interface{}),
		Validator:      validator.New(),
	}
}

// Reset the context.
func (c *Context) Reset(w http.ResponseWriter, r *http.Request) {
	c.responseWriter = w
	c.request = r
	c.store = make(map[string]interface{})
	c.LastError = nil
}

func isJson(s string) bool {
	s = strings.Replace(strings.ToUpper(s), " ", "", -1)
	j := strings.Replace(strings.ToUpper(MIMEApplicationJSON), " ", "", -1)
	jsonCharset := strings.Replace(strings.ToUpper(MIMEApplicationJSONCharsetUTF8), " ", "", -1)

	return strings.EqualFold(s, j) || strings.EqualFold(s, jsonCharset)
}

// JSONBody parses the JSON request body.
func (c *Context) JSONBody(v interface{}) error {
	if c.request.ContentLength == 0 {
		return errNoBody
	}

	if conType := c.request.Header.Get(HeaderContentType); !isJson(conType) {
		return errNotJSONBody
	}

	return json.NewDecoder(c.request.Body).Decode(v)
}

// ServeJSON sends a JSON response.
func (c *Context) ServeJSON(v interface{}) error {
	if v == nil {
		return errEmptyResponse
	}

	resp, err := json.Marshal(v)
	if err != nil {
		return err
	}

	c.responseWriter.Header().Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	_, err = c.responseWriter.Write(resp)
	return err
}

// Redirect does redirection to localurl with http header status code.
func (c *Context) Redirect(status int, url string) error {
	if status < 300 || status > 308 {
		return errInvalidRedirectCode
	}
	c.responseWriter.Header().Set(HeaderLocation, url)
	c.responseWriter.WriteHeader(status)
	return nil
}

// Cookies return all cookies.
func (c *Context) Cookies() []*http.Cookie {
	return c.request.Cookies()
}

// GetCookie Get cookie from request by a given key.
func (c *Context) GetCookie(key string) (*http.Cookie, error) {
	return c.request.Cookie(key)
}

// SetCookie Set cookie for response.
func (c *Context) SetCookie(name string, value string) {
	cook := http.Cookie{
		Name:  name,
		Value: value,
	}
	http.SetCookie(c.responseWriter, &cook)
}

// Request return the request.
func (c *Context) Request() *http.Request {
	return c.request
}

// SetRequest set the r as the new request.
func (c *Context) SetRequest(r *http.Request) {
	c.request = r
}

// Response return the responseWriter
func (c *Context) Response() http.ResponseWriter {
	return c.responseWriter
}

// FormValue returns the first value for the named component of the query.
func (c *Context) FormValue(name string) string {
	return c.request.FormValue(name)
}

// FormParams return the parsed form data
func (c *Context) FormParams() (url.Values, error) {
	if strings.HasPrefix(c.request.Header.Get(HeaderContentType), MIMEMultipartForm) {
		if err := c.request.ParseMultipartForm(defaultMemory); err != nil {
			return nil, err
		}
	} else {
		if err := c.request.ParseForm(); err != nil {
			return nil, err
		}
	}
	return c.request.Form, nil
}

// SetHeader Set header for response.
func (c *Context) SetHeader(key, val string) {
	c.responseWriter.Header().Set(key, val)
}

// WriteHeader sends an HTTP response header with status code.
func (c *Context) WriteHeader(code int) error {
	c.responseWriter.WriteHeader(code)
	return nil
}

// GetHeader Get header from request by a given key.
func (c *Context) GetHeader(key string) string {
	return c.request.Header.Get(key)
}

// Validate the parameters.
func (c *Context) Validate(val interface{}) error {
	return c.Validator.Struct(val)
}

// Set a kev/value on context.
func (c *Context) Set(key string, value interface{}) {
	c.store[key] = value
}

// Get a value for a key.
func (c *Context) Get(key string) interface{} {
	return c.store[key]
}
