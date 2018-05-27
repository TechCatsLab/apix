/*
 * Revision History:
 *     Initial: 2018/5/27        ShiChao
 */

package server

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// Router register routes to be matched and dispatched to a handler.
type Router struct {
	router     *mux.Router
	ctxPool    sync.Pool
	errHandler func(*Context)
}

// NewRouter returns a router.
func NewRouter() *Router {
	r := &Router{
		router:     mux.NewRouter(),
		errHandler: func(_ *Context) {},
	}

	r.ctxPool.New = func() interface{} {
		return NewContext(nil, nil)
	}

	r.router.NotFoundHandler = http.NotFoundHandler()
	r.router.MethodNotAllowedHandler = MethodNotAllowedHandler()

	return r
}

// Handler returns a http.Handler.
func (rt *Router) Handler() http.Handler {
	return rt.router
}

// SetErrorHandler attach a global error handler on router.
func (rt *Router) SetErrorHandler(h func(*Context)) {
	rt.errHandler = h
}

// Get adds a route path access via GET method.
func (rt *Router) Get(pattern string, handler HandlerFunc, filters ...FilterFunc) {
	rt.router.HandleFunc(pattern, rt.wrapHandlerFunc(handler, filters...)).Methods("GET")
}

// Post adds a route path access via POST method.
func (rt *Router) Post(pattern string, handler HandlerFunc, filters ...FilterFunc) {
	rt.router.HandleFunc(pattern, rt.wrapHandlerFunc(handler, filters...)).Methods("POST")
}

// Wraps a HandlerFunc to a http.HandlerFunc.
func (rt *Router) wrapHandlerFunc(f HandlerFunc, filters ...FilterFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := rt.ctxPool.Get().(*Context)
		defer rt.ctxPool.Put(c)
		c.Reset(w, r)

		if len(filters) > 0 {
			for _, filter := range filters {
				if passed := filter(c); !passed {
					c.LastError = errFilterNotPassed
					return
				}
			}
		}

		if err := f(c); err != nil {
			c.LastError = err
			rt.errHandler(c)
		}
	}
}

// MethodNotAllowedHandler returns a simple request handler
// that replies to each request with a ``405 method not allowed'' reply.
func MethodNotAllowedHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
	})
}
