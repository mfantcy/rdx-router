package mux

import "net/http"

type methodContext struct {
	handler    http.Handler
	handleFunc http.HandlerFunc
}

func (mc *methodContext) Use(middleware ...MiddlewareFunc) {
	mc.handleFunc = mc.handler.ServeHTTP
	for _, m := range middleware {
		mc.handleFunc = m(mc.handleFunc).ServeHTTP
	}
}

type Route map[string]*methodContext

func (r Route) Methods() (methods []string) {
	for key := range r {
		methods = append(methods, key)
	}
	return
}

func (r Route) MethodHandleFunc(method string) (handleFunc http.HandlerFunc) {
	if ctx, ok := r[method]; ok {
		handleFunc = ctx.handleFunc
	}
	return handleFunc
}
