package rdx_router

import (
	"github.com/mfantcy/rdx-router/tree"
)

type params tree.Pairs

func toParams(p tree.Pairs) params {
	return params(p)
}

func (p params) ValueOf(paramName string) string {
	for _, param := range p {
		if param.Key == paramName {
			return param.Value
		}
	}
	return ""
}

func (p params) Value(index int) string {
	if index < p.Count() {
		return p[index].Value
	}
	return ""
}

func (p params) Count() int {
	return len(p)
}
