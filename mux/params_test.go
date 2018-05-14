package mux

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mfantcy/rdx-router/tree"
)

func TestParams_newParams(t *testing.T) {
	var p []*tree.Pair
	ps := newParams(p)
	assert.IsType(t, (*Params)(nil), ps)
}

func TestParams_ValueOf(t *testing.T) {
	var p []*tree.Pair
	ps := newParams(p)
	assert.Empty(t, ps.ValueOf("key"))

	p2 := []*tree.Pair{&tree.Pair{"abc", "cde"}, &tree.Pair{"123", "432"}}
	ps2 := newParams(p2)
	assert.Equal(t, "cde", ps2.ValueOf("abc"))
	assert.Equal(t, "432", ps2.ValueOf("123"))
	assert.Equal(t, "", ps2.ValueOf("not_exist"))
}

func TestParams_Value(t *testing.T) {
	var p []*tree.Pair
	ps := newParams(p)
	assert.Empty(t, ps.ValueOf("key"))

	p2 := []*tree.Pair{&tree.Pair{"abc", "cde"}, &tree.Pair{"123", "432"}}
	ps2 := newParams(p2)
	assert.Equal(t, "cde", ps2.Value(0))
	assert.Equal(t, "432", ps2.Value(1))
	assert.Equal(t, "", ps2.Value(3))
}

func TestParams_Count(t *testing.T) {
	var p []*tree.Pair
	ps := newParams(p)
	assert.Equal(t, 0, ps.Count())
	p2 := []*tree.Pair{&tree.Pair{"abc", "cde"}, &tree.Pair{"123", "432"}}
	ps2 := newParams(p2)
	assert.Equal(t, 2, ps2.Count())
}
