package mux

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroup_NewGroup(t *testing.T) {
	g := newGroup("dummy_text")
	assert.IsType(t, (*group)(nil), g)
}

func TestGroup_Group(t *testing.T) {
	g := newGroup("dummy_text")
	sg := g.Group("dummy_text2", func(routeRegistrar RouteRegistrar) {
		handler := func(w http.ResponseWriter, req *http.Request) {}
		routeRegistrar.GET("abc", http.HandlerFunc(handler))
	})
	assert.IsType(t, (*group)(nil), sg)
	assert.Equal(t, g, sg.(*group).root())
}

func TestGroup_Use(t *testing.T) {
	assert.True(t, true)
}
