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
	"net/url"
	"strings"
	"time"
)

// A Client is an HTTP client.
type Client struct {
	*http.Client
}

// NewClient returns an HTTP client.
func NewClient(client *http.Client) *Client {
	if client != nil {
		return &Client{client}
	}

	return &Client{&http.Client{}}
}

// NewClientWithProxy -
func NewClientWithProxy(proxy string) (*Client, error) {
	if proxy == "" {
		return &Client{&http.Client{}}, nil
	}

	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return nil, err
	}

	return &Client{&http.Client{Transport: &http.Transport{
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
	}}}, nil
}

// Do sends an HTTP request and returns an HTTP response.
func (c *Client) Do(req *Request) (*Response, error) {
	resp, err := c.Client.Do(req.Request)
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
func (c *Client) GetFile(url, target string) (int64, error) {
	resp, err := c.Get(url)
	if err != nil {
		return 0, err
	}

	return saveFile(resp, target)
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
