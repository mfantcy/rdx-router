package mux

import (
	"net/http"
)

type ParamsHolder interface {
	ValueOf(paramName string) string
	Value(index int) string
	Count() int
}

type MiddlewareFunc func(next http.Handler) http.Handler

type MiddlewareRegistrar interface {
	Use(middleware ...MiddlewareFunc)
}

type RouteRegistrar interface {
	Handle(path string, handleFunc http.Handler, httpMethod ...string) MiddlewareRegistrar

	GET(path string, handleFunc http.Handler) MiddlewareRegistrar
	POST(path string, handleFunc http.Handler) MiddlewareRegistrar
	PUT(path string, handleFunc http.Handler) MiddlewareRegistrar
	DELETE(path string, handleFunc http.Handler) MiddlewareRegistrar
	OPTIONS(path string, handleFunc http.Handler) MiddlewareRegistrar
	HEAD(path string, handleFunc http.Handler) MiddlewareRegistrar
	PATCH(path string, handleFunc http.Handler) MiddlewareRegistrar

	Group(path string, groupFunc func(routeRegistrar RouteRegistrar)) MiddlewareRegistrar
}

type RouteHandler interface {
	Methods() []string
	MethodHandleFunc(method string) http.HandlerFunc
}
