package mux

import (
	"context"
	"github.com/mfantcy/rdx-router/tree"
	"net/http"
)

const ParamCtxKey = "RequestParams"

type Params struct {
	pairs []*tree.Pair
}

func (p *Params) ValueOf(paramName string) string {
	for _, pair := range p.pairs {
		if pair.Name == paramName {
			return pair.Value
		}
	}
	return ""
}

func (p *Params) Value(index int) string {
	if index < p.Count() {
		return p.pairs[index].Value
	}
	return ""
}

func RequestParams(r *http.Request) ParamsHolder {
	requestParams := r.Context().Value(ParamCtxKey)
	if rp, ok := requestParams.(ParamsHolder); ok {
		return rp
	}
	return &Params{}
}

func (p *Params) Count() int {
	return len(p.pairs)
}

func toWithRequestParams(r *http.Request, params ParamsHolder) *http.Request {
	ctx := context.WithValue(r.Context(), ParamCtxKey, params)
	return r.WithContext(ctx)
}

func newParams(pairs []*tree.Pair) *Params {
	return &Params{pairs}
}
