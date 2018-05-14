package mux

import "net/http"

type group struct {
	parent          *group
	subGroups       []*group
	path            string
	methodCxtRefs   []*methodContext
	routes          map[string]map[string]*methodContext
	middlewareChain []MiddlewareFunc
}

func newGroup(path string) *group {
	return &group{path: path, routes: make(map[string]map[string]*methodContext)}
}

func (g *group) root() (group *group) {
	group = g
	for group.parent != nil {
		group = group.parent
	}
	return
}

func (g *group) getRoutes() (routes map[string]map[string]*methodContext) {
	routes = make(map[string]map[string]*methodContext)
	for p, mctx := range g.routes {
		routes[g.path+p] = mctx
	}
	for _, subGroup := range g.subGroups {
		subRoutes := subGroup.getRoutes()
		for p, mctx := range subRoutes {
			routes[g.path+p] = mctx
		}
	}
	return
}

func (g *group) Use(middleware ...MiddlewareFunc) {
	g.middlewareChain = middleware
}

func (g *group) Handle(path string, handleFunc http.Handler, httpMethod ...string) MiddlewareRegistrar {
	methodCtx := &methodContext{handleFunc, handleFunc.ServeHTTP}
	mctx, ok := g.routes[path]
	if !ok {
		mctx = make(map[string]*methodContext)
		g.routes[path] = mctx
	}
	for _, m := range httpMethod {
		mctx[m] = methodCtx
	}
	g.methodCxtRefs = append(g.methodCxtRefs, methodCtx)
	return methodCtx
}

func (g *group) GET(path string, handleFunc http.Handler) MiddlewareRegistrar {
	return g.Handle(path, handleFunc, "GET")
}

func (g *group) POST(path string, handleFunc http.Handler) MiddlewareRegistrar {
	return g.Handle(path, handleFunc, "POST")
}

func (g *group) PUT(path string, handleFunc http.Handler) MiddlewareRegistrar {
	return g.Handle(path, handleFunc, "PUT")
}

func (g *group) DELETE(path string, handleFunc http.Handler) MiddlewareRegistrar {
	return g.Handle(path, handleFunc, "DELETE")
}

func (g *group) OPTIONS(path string, handleFunc http.Handler) MiddlewareRegistrar {
	return g.Handle(path, handleFunc, "OPTIONS")
}

func (g *group) HEAD(path string, handleFunc http.Handler) MiddlewareRegistrar {
	return g.Handle(path, handleFunc, "HEAD")
}

func (g *group) PATCH(path string, handleFunc http.Handler) MiddlewareRegistrar {
	return g.Handle(path, handleFunc, "PATCH")
}

func (g *group) Group(path string, groupFunc func(routeRegistrar RouteRegistrar)) MiddlewareRegistrar {
	subGroup := newGroup(path)
	subGroup.parent = g
	groupFunc(subGroup)
	g.subGroups = append(g.subGroups, subGroup)
	return subGroup
}
