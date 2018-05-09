package rdx_router

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mfantcy/rdx-router/tree"
)

func TestParams_ToParams(t *testing.T) {
	p := tree.Pairs{}
	ps := toParams(p)
	assert.IsType(t, (params)(nil), ps)
}

func TestParams_ValueOf(t *testing.T) {
	p := tree.Pairs{}
	ps := toParams(p)
	assert.Empty(t, ps.ValueOf("key"))

	p2 := tree.Pairs{&tree.Pair{"abc", "cde"}, &tree.Pair{"123", "432"}}
	ps2 := toParams(p2)
	assert.Equal(t, "cde", ps2.ValueOf("abc"))
	assert.Equal(t, "432", ps2.ValueOf("123"))
	assert.Equal(t, "", ps2.ValueOf("not_exist"))
}

func TestParams_Value(t *testing.T) {
	p := tree.Pairs{}
	ps := toParams(p)
	assert.Empty(t, ps.Value(2))

	p2 := tree.Pairs{&tree.Pair{"abc", "cde"}, &tree.Pair{"123", "432"}}
	ps2 := toParams(p2)
	assert.Equal(t, "cde", ps2.Value(0))
	assert.Equal(t, "432", ps2.Value(1))
	assert.Equal(t, "", ps2.Value(3))
}

func TestParams_Count(t *testing.T) {
	p := tree.Pairs{}
	ps := toParams(p)
	assert.Equal(t, 0, ps.Count())
	p2 := tree.Pairs{&tree.Pair{"abc", "cde"}, &tree.Pair{"123", "432"}}
	ps2 := toParams(p2)
	assert.Equal(t, 2, ps2.Count())
}
