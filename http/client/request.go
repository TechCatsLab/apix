/*
 * Revision History:
 *     Initial: 2018/05/31        Li Zebang
 */

package client

import (
	"io"
	"net/http"
)

const (
	HeaderAuthorization = "Authorization"

	AuthSchemeBearer = "Bearer"
)

// A Request represents an HTTP request to be sent by a client.
type Request struct {
	HTTPRequest *http.Request
}

// NewRequest returns a new Request given a method, URL, and optional body.
func NewRequest(method, url string, body io.Reader) (*Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	return &Request{req}, nil
}

// AddHeader adds the key, value pair to the header.
// It appends to any existing values associated with key.
func (r *Request) AddHeader(key, value string) {
	r.HTTPRequest.Header.Add(key, value)
}

// AddHeaders adds the multiple headers.
func (r *Request) AddHeaders(headers map[string]string) {
	for key, value := range headers {
		r.HTTPRequest.Header.Add(key, value)
	}
}

// DelHeader deletes the values associated with key.
func (r *Request) DelHeader(key string) {
	r.HTTPRequest.Header.Del(key)
}

// GetHeader gets the first value associated with the given key.
// If there are no values associated with the key, Get returns "".
func (r *Request) GetHeader(key, value string) string {
	return r.HTTPRequest.Header.Get(key)
}

// SetHeader sets the header entries associated with key to the single
// element value. It replaces any existing values associated with key.
func (r *Request) SetHeader(key, value string) {
	r.HTTPRequest.Header.Set(key, value)
}

// SetToken -
func (r *Request) SetToken(token string) {
	r.HTTPRequest.Header.Set(HeaderAuthorization, AuthSchemeBearer+" "+token)
}
