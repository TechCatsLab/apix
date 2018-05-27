/*
 * Revision History:
 *     Initial: 2018/05/25        ShiChao
 */

package xhttp

import (
	"net"
	"net/http"
	"fmt"
	stdCtx "context"
	"sync"
	"time"
	"github.com/urfave/negroni"
	"github.com/gorilla/mux"
)

type HandleFunc func(ctx Context) error

type Server struct {
	listener    net.Listener
	server      *http.Server
	router      *mux.Router
	middlewares []negroni.Handler
	pool        sync.Pool
	errHandler  HandleFunc
	stop        chan bool
}

func New() (server *Server) {
	server = &Server{
		server: new(http.Server),
		router: &mux.Router{},
		stop:   make(chan bool),
	}

	server.pool.New = func() interface{} {
		return newContext()
	}

	return
}

func defaultErrHandler(ctx Context) error {
	// todo: add log
	res := ctx.Response()
	http.Error(res, "405 method not allowed", http.StatusMethodNotAllowed)
	return nil
}

func (server *Server) GetContext() Context {
	return server.pool.Get().(Context)
}

func (server *Server) ReleaseContext(ctx context) {
	server.pool.Put(ctx)
}

func (server *Server) Use(fn negroni.Handler) {
	server.middlewares = append(server.middlewares, fn)
}

func (server *Server) wrapHandlerFunc(f HandleFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := server.pool.Get().(Context)
		ctx.Reset(r, w)
		defer server.pool.Put(ctx)

		if err := f(ctx); err != nil {
			if h := server.errHandler; h != nil {
				h(ctx)
			} else {
				server.errHandler = defaultErrHandler
				defaultErrHandler(ctx)
			}
		}
	}
}

func MethodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
	})
}

func (server *Server) Get(path string, f HandleFunc) {
	server.router.HandleFunc(path, server.wrapHandlerFunc(f)).Methods(GET)
}

func (server *Server) Post(path string, f HandleFunc) {
	server.router.HandleFunc(path, server.wrapHandlerFunc(f)).Methods(POST)
}

func (server *Server) init() {
	server.router.NotFoundHandler = http.NotFoundHandler()
	server.router.MethodNotAllowedHandler = MethodNotAllowedHandler()

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	for _, m := range server.middlewares {
		n.Use(m)
	}

	n.UseHandler(server.router)

	server.server.Handler = n
}

func (server *Server) Run(addr string) {
	l, err := net.Listen(TCP, addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.listener = l

	server.init()

	err = server.server.Serve(l)
	fmt.Println(err)
}

func (server *Server) RunTLS(addr string, certFile, keyFile string) {
	l, err := net.Listen(TCP, addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.listener = l

	server.init()

	err = server.server.ServeTLS(l, certFile, keyFile)
	fmt.Println(err)
}

func (server *Server) GracefulClose() {
	ctx, cancel := stdCtx.WithTimeout(stdCtx.Background(), 5*time.Second)

	if err := server.server.Shutdown(ctx); err != nil {
		server.server.Close()
	}
	cancel()
}
