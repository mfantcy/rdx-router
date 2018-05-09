package rdx_router

type methodContext struct {
	handler    HandleFunc
	handleFunc HandleFunc
}

func (mc *methodContext) Use(middleware ...MiddlewareFunc) {
	mc.handleFunc = mc.handler
	for _, m := range middleware {
		mc.handleFunc = m(mc.handleFunc)
	}
}

type Route map[string]*methodContext

func (r Route) Methods() (methods []string) {
	for key := range r {
		methods = append(methods, key)
	}
	return
}

func (r Route) MethodHandleFunc(method string) (handleFunc HandleFunc) {
	if ctx, ok := r[method]; ok {
		handleFunc = ctx.handleFunc
	}
	return handleFunc
}
