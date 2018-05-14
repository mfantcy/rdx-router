package mux

import (
	"net/http"
	"regexp"

	"github.com/mfantcy/rdx-router/tree"
	"strings"
)

type PanicHandleFunc func(recovered interface{}) http.HandlerFunc

type Router struct {
	FixTrailingSlash bool

	HandleMethodNotAllowed bool

	HandleOPTIONS bool

	NotFoundHandler http.HandlerFunc

	MethodNotAllowedHandler http.HandlerFunc

	PanicFunc PanicHandleFunc

	tree tree.TrieInterface

	middlewareChain []MiddlewareFunc
}

func (r *Router) Use(middleware ...MiddlewareFunc) {
	r.middlewareChain = middleware
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.PanicFunc != nil {
		defer r.recover(w, req)
	}
	var handleFunc http.HandlerFunc
	if rt, p, ok := r.tree.Lookup(req.URL.Path, r.FixTrailingSlash); ok && rt != nil { //resource found
		route := rt.(Route)
		handleFunc = route.MethodHandleFunc(req.Method)
		if len(p) > 0 {
			req = toWithRequestParams(req, newParams(p))
		}
		if handleFunc == nil {
			if req.Method == "OPTIONS" && r.HandleOPTIONS {
				handleFunc = func(w http.ResponseWriter, req *http.Request) {
					allowedMethods := route.Methods()
					allowedMethods = uniqueAppend(allowedMethods, "OPTIONS")
					w.Header().Set("Allow", strings.Join(allowedMethods, " "))
					w.WriteHeader(200)
				}
			} else if r.MethodNotAllowedHandler != nil || r.HandleMethodNotAllowed { //method not allowed
				handleFunc = func(w http.ResponseWriter, req *http.Request) {
					allowedMethods := route.Methods()
					if r.HandleOPTIONS {
						allowedMethods = uniqueAppend(allowedMethods, "OPTIONS")
					}
					w.Header().Set("Allow", strings.Join(allowedMethods, " "))
					if r.MethodNotAllowedHandler != nil {
						r.MethodNotAllowedHandler.ServeHTTP(w, req)
					} else {
						http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
					}
				}
			}
		}
	}

	//not found
	if handleFunc == nil {
		handleFunc = func(w http.ResponseWriter, req *http.Request) {
			if r.NotFoundHandler != nil {
				r.NotFoundHandler.ServeHTTP(w, req)
			} else {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			}
		}

	}
	//global middleware
	for _, middlewareFunc := range r.middlewareChain {
		handleFunc = middlewareFunc(handleFunc).ServeHTTP
	}
	handleFunc(w, req)
	return
}

func NewRouter() *Router {
	return &Router{
		tree:                   tree.NewTree(),
		FixTrailingSlash:       true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
	}
}

func (r *Router) Handle(path string, handler http.Handler, httpMethod ...string) MiddlewareRegistrar {
	methodCxt := &methodContext{handler, handler.ServeHTTP}
	r.handle(path, methodCxt, httpMethod...)
	return methodCxt
}

func (r *Router) GET(path string, handler http.Handler) MiddlewareRegistrar {
	return r.Handle(path, handler, "GET")
}

func (r *Router) POST(path string, handler http.Handler) MiddlewareRegistrar {
	return r.Handle(path, handler, "POST")
}

func (r *Router) PUT(path string, handler http.Handler) MiddlewareRegistrar {
	return r.Handle(path, handler, "PUT")
}

func (r *Router) DELETE(path string, handler http.Handler) MiddlewareRegistrar {
	return r.Handle(path, handler, "DELETE")
}

func (r *Router) OPTIONS(path string, handler http.Handler) MiddlewareRegistrar {
	return r.Handle(path, handler, "OPTIONS")
}

func (r *Router) HEAD(path string, handler http.Handler) MiddlewareRegistrar {
	return r.Handle(path, handler, "HEAD")
}

func (r *Router) PATCH(path string, handler http.Handler) MiddlewareRegistrar {
	return r.Handle(path, handler, "PATCH")
}

func (r *Router) Group(path string, groupFunc func(routeRegistrar RouteRegistrar)) MiddlewareRegistrar {
	group := newGroup(path)
	groupFunc(group)
	routes := group.root().getRoutes()
	for path, m := range routes {
		for method, methodCtx := range m {
			r.handle(path, methodCtx, method)
		}
	}
	return group
}

func (r *Router) handle(path string, methodCtx *methodContext, httpMethod ...string) {
	r.tree.AddThen(path, func(context interface{}) interface{} {
		var route Route
		if r, ok := context.(Route); ok {
			route = r
		} else {
			route = make(Route)
		}
		for _, m := range httpMethod {
			if matches, err := regexp.MatchString("^[A-Z]+(-[A-Z]+)*$", m); !matches || err != nil {
				panic("http method '" + m + "' is not valid")
			}
			route[m] = methodCtx
		}
		return route
	})
}

func uniqueAppend(a []string, s string) []string {
	for _, m := range a {
		if m == s {
			return a
		}
	}
	a = append(a, s)
	return a
}

func (r *Router) recover(w http.ResponseWriter, req *http.Request) {
	if rev := recover(); rev != nil {
		r.PanicFunc(rev)(w, req)
	}
}
