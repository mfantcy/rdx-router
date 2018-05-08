package rdx_router

import (
	"net/http"
)

type MiddlewareRegistrar interface {
	Use(middleware ...MiddlewareFunc) MiddlewareRegistrar
}

type RouteRegistrar interface {
	Handle(path string, handleFunc HandleFunc, httpMethod ...string) MiddlewareRegistrar

	GET(path string, handleFunc HandleFunc) MiddlewareRegistrar
	POST(path string, handleFunc HandleFunc) MiddlewareRegistrar
	PUT(path string, handleFunc HandleFunc) MiddlewareRegistrar
	DELETE(path string, handleFunc HandleFunc) MiddlewareRegistrar
	OPTIONS(path string, handleFunc HandleFunc) MiddlewareRegistrar
	HEAD(path string, handleFunc HandleFunc) MiddlewareRegistrar
	PATCH(path string, handleFunc HandleFunc) MiddlewareRegistrar

	Group(path string, groupFunc func(routeRegistrar RouteRegistrar)) MiddlewareRegistrar
}

type RouteHandler interface {
	Methods() []string
	MethodHandleFunc(method string) HandleFunc
}

type ParamsHolder interface {
	ValueOf(paramName string) string
	Value(index int) string
	Count() int
}

type MiddlewareFunc func(next HandleFunc) HandleFunc

type HandleFunc func(w http.ResponseWriter, req *http.Request, params ParamsHolder)

