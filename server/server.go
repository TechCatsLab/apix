/*
 * Revision History:
 *     Initial: 2018/5/26        ShiChao
 */

package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"os"
	"time"
	"fmt"

	"github.com/urfave/negroni"
)

const (
	// charset
	charsetUTF8 = "charset=UTF-8"

	// HTTP methods
	PUT     = "PUT"
	GET     = "GET"
	POST    = "POST"
	HEAD    = "HEAD"
	PATCH   = "PATCH"
	DELETE  = "DELETE"
	OPTIONS = "OPTIONS"

	// MIME
	MIMEApplicationJSON            = "application/json"
	MIMEApplicationJSONCharsetUTF8 = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEMultipartForm              = "multipart/form-data"

	// Headers
	HeaderOrigin        = "Origin"
	HeaderAccept        = "Accept"
	HeaderVary          = "Vary"
	HeaderCookie        = "Cookie"
	HeaderSetCookie     = "Set-Cookie"
	HeaderUpgrade       = "Upgrade"
	HeaderContentType   = "Content-Type"
	HeaderLocation      = "Location"
	HeaderAuthorization = "Authorization"

	// Access control
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"
)

var (
	errNoRouter  = errors.New("entrypoint requires a router")
	errTLSConfig = errors.New("cert or key file in the TLS configuration does not exist")
)

// Entrypoint represents a http server.
type Entrypoint struct {
	configuration *Configuration
	tlsConfig     *TLSConfiguration
	server        *http.Server
	listener      net.Listener
	middlewares   []negroni.Handler
	stop          chan bool
	signals       chan os.Signal
}

// NewEntrypoint creates a new Entrypoint.
func NewEntrypoint(conf *Configuration, tlsConf *TLSConfiguration) *Entrypoint {
	return &Entrypoint{
		configuration: conf,
		tlsConfig:     tlsConf,
		stop:          make(chan bool, 1),
		signals:       make(chan os.Signal, 1),
		middlewares:   []negroni.Handler{},
	}
}

// Prepare the entrypoint for serving requests.
func (ep *Entrypoint) prepare(router http.Handler) error {
	var (
		err       error
		tlsConfig *tls.Config
		listener  net.Listener
	)

	if tlsConfig, err = ep.createTLSConfig(); err != nil {
		return err
	}

	listener, err = net.Listen("tcp", ep.configuration.Address)
	if err != nil {
		return err
	}

	ep.listener = listener
	ep.server = &http.Server{
		Addr:      ep.configuration.Address,
		Handler:   ep.buildRouter(router),
		TLSConfig: tlsConfig,
	}

	return nil
}

// Create the TLS Configuration for the http server.
func (ep *Entrypoint) createTLSConfig() (*tls.Config, error) {
	var (
		err    error
		exists bool
	)

	if ep.tlsConfig == nil {
		return nil, nil
	}

	if exists, err = FileExists(ep.tlsConfig.Cert); !exists {
		return nil, errTLSConfig
	}

	if exists, err = FileExists(ep.tlsConfig.Key); !exists {
		return nil, errTLSConfig
	}

	cert, err := tls.LoadX509KeyPair(ep.tlsConfig.Cert, ep.tlsConfig.Key)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	return config, nil
}

func FileExists(filename string) (bool, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (ep *Entrypoint) buildRouter(router http.Handler) http.Handler {
	n := negroni.New()

	for _, mw := range ep.middlewares {
		n.Use(mw)
	}

	n.Use(negroni.Wrap(http.HandlerFunc(router.ServeHTTP)))

	return n
}

func (ep *Entrypoint) startServer() error {
	if ep.tlsConfig != nil {
		return ep.server.ServeTLS(ep.listener, "", "")
	}

	return ep.server.Serve(ep.listener)
}

// AttachMiddleware attach a new middleware on entrypoint.
func (ep *Entrypoint) AttachMiddleware(handler negroni.Handler) {
	ep.middlewares = append(ep.middlewares, handler)
}

// Start the entrypoint.
func (ep *Entrypoint) Start(router http.Handler) error {
	if router == nil {
		return errNoRouter
	}

	if err := ep.prepare(router); err != nil {
		return err
	}

	ep.configureSignals()

	go ep.listenSignals()
	go ep.startServer()

	fmt.Println("Serving on:", ep.configuration.Address)

	return nil
}

// Wait until stop channel emits a value.
func (ep *Entrypoint) Run() {
	<-ep.stop
}

// Stop the http server.
func (ep *Entrypoint) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// graceful shutdown
	if err := ep.server.Shutdown(ctx); err != nil {
		ep.server.Close()
	}
	cancel()

	close(ep.stop)
}
