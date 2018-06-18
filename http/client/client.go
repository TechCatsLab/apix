/*
 * Revision History:
 *     Initial: 2018/05/28        Li Zebang
 */

package client

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// A Client is an HTTP client.
type Client struct {
	HTTPClient *http.Client
	Headers    map[string]string
}

// DefaultTransport -
func DefaultTransport() http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// DefaultTimeout -
const DefaultTimeout = 30 * time.Second

// NewClient returns an HTTP client.
func NewClient(transport http.RoundTripper, jar http.CookieJar, timeout time.Duration) (*Client, error) {
	var client http.Client

	if transport != nil {
		client.Transport = transport
	} else {
		client.Transport = DefaultTransport()
	}

	if jar != nil {
		client.Jar = jar
	} else {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, err
		}
		client.Jar = jar
	}

	if timeout > 0 {
		client.Timeout = timeout
	} else {
		client.Timeout = DefaultTimeout
	}

	return &Client{HTTPClient: &client, Headers: make(map[string]string)}, nil
}

// NewClientWithProxy -
func NewClientWithProxy(proxy string) (*Client, error) {
	if proxy == "" {
		return NewClient(nil, nil, 0)
	}

	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return NewClient(transport, nil, 0)
}

// SetToken -
func (c *Client) SetToken(token string) {
	c.Headers[HeaderAuthorization] = AuthSchemeBearer + " " + token
}

// Do sends an HTTP request and returns an HTTP response.
func (c *Client) Do(req *Request) (*Response, error) {
	req.AddHeaders(c.Headers)

	resp, err := c.HTTPClient.Do(req.HTTPRequest)
	if err != nil {
		return nil, err
	}

	return &Response{resp}, err
}

// Get sends a Get HTTP request and returns an HTTP response.
func (c *Client) Get(url string) (*Response, error) {
	req, err := NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

// GetFile sends a Get HTTP request and save the file.
func (c *Client) GetFile(url, directory string) (int64, error) {
	resp, err := c.Get(url)
	if err != nil {
		return 0, err
	}

	return resp.SaveAsFile(directory)
}

// Post sends a Post HTTP request and returns an HTTP response.
func (c *Client) Post(url string, contentType string, body io.Reader) (*Response, error) {
	req, err := NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.SetHeader("Content-Type", contentType)

	return c.Do(req)
}

// PostForm sends a Post HTTP request with form data and returns an HTTP response.
func (c *Client) PostForm(url string, data url.Values) (*Response, error) {
	return c.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

// PostJSON sends a Post HTTP request with JSON data and returns an HTTP response.
func (c *Client) PostJSON(url string, body interface{}) (*Response, error) {
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return c.Post(url, "application/json", strings.NewReader(string(reqBody)))
}
