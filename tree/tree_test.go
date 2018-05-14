package tree

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func dumpNode(n *node, prefix string) {
	fmt.Printf("%s%s fullpath:%s %v \r\n", prefix, n.path, n.FullPathPattern(), n.leaf)
	for l := len(n.path); l > 0; l-- {
		prefix += " "
	}
	for _, child := range n.staticBranches {
		if child != nil {
			dumpNode(child, prefix)
		}
	}
	for _, regexpChild := range n.regexpBranches {
		dumpNode(regexpChild, prefix)
	}
	if n.wildBranch != nil {
		dumpNode(n.wildBranch, prefix)
	}
}

func pairsString(pairs []*Pair) string {
	res := ""
	for _, pairp := range pairs {
		pstr := fmt.Sprintf("%v", pairp)
		if res != "" {
			res += ", " + pstr
		} else {
			res += pstr
		}
	}
	return "Pairs{" + res + "}"
}

func assertAddExceptFullPath(t *testing.T, tree *node, pattern string, context interface{}, expected string, params ...[]string) {
	edge := tree.Add(pattern, context)
	assert.Equal(t, expected, edge.FullPathPattern())
	assert.Equal(t, context, edge.Context())
	if len(params) > 0 {
		if p := edge.Params(); len(params[0]) == len(p) {
			for idx, param := range params[0] {
				if p[idx] != param {
					assert.Fail(t, fmt.Sprintf("expected: %v actual:%v", params, p))
					break
				}
			}
		} else {
			assert.Fail(t, fmt.Sprintf("expected: %v actual:%v", params, p))
		}
	}
}

func assertFound(t *testing.T, tree *node, lookup string, fixTailingSlash bool, expectedContext interface{}) {
	ctx, _, ok := tree.Lookup(lookup, fixTailingSlash)
	if !ok {
		assert.Fail(t, fmt.Sprintf("Looking up \"%s\" not found", lookup))
	}
	if expectedContext != nil {
		assert.Equal(t, expectedContext, ctx)
	}
}

func assertAddPanic(t *testing.T, tree *node, pattern string) {
	assert.Panics(t, func() { tree.Add(pattern, nil) })
}

func assertNotFound(t *testing.T, tree *node, lookup string, fixTailingSlash bool) {
	_, _, ok := tree.Lookup(lookup, fixTailingSlash)
	if ok {
		assert.Fail(t, fmt.Sprintf("Looking up \"%s\" should not found", lookup))
	}
}

func assertFoundParams(t *testing.T, tree *node, lookup string, fixTailingSlash bool, expectedParams []*Pair, expectedContext interface{}) {
	ctx, p, ok := tree.Lookup(lookup, fixTailingSlash)
	if !ok {
		assert.Fail(t, lookup+" not found")
		return
	}
	if expectedParams != nil {
		if len(p) != len(expectedParams) {
			assert.Fail(t, "expected: "+pairsString(expectedParams)+" actual: "+pairsString(p))
		} else {
			for _, v := range expectedParams {
				found := false
				for _, pv := range p {
					//fmt.Println(pv)
					if pv != nil && v.Name == pv.Name && v.Value == pv.Value {
						found = true
					}
				}
				if !found {
					assert.Fail(t, fmt.Sprintf("expected &Pair{Name: \"%s\", Value: \"%s\"} not found ", v.Name, v.Value)+" expected: "+pairsString(expectedParams)+" actual: "+pairsString(p))
				}
			}
		}
	}
	if expectedContext != nil {
		assert.Equal(t, expectedContext, ctx)
	}
}

func TestSplitEdge(t *testing.T) {
	tree := NewTree()
	assertAddExceptFullPath(t, tree, "/path/1234", "1", "/path/1234", []string{})
	assertAddExceptFullPath(t, tree, "/path/{:123}", "2", "/path/{:123}", []string{""})
	assertAddExceptFullPath(t, tree, "/path/{:cde}", "3", "/path/{:cde}", []string{""})
	assertAddExceptFullPath(t, tree, "/path/{param}", "4", "/path/{param}", []string{"param"})
	assertAddExceptFullPath(t, tree, "/pathto/{:123}", 5, "/pathto/{:123}", []string{""})

	assertFound(t, tree, "/path/1234", false, "1")
	assertFound(t, tree, "/path/123", false, "2")
	assertFound(t, tree, "/path/cde", false, "3")
	assertFound(t, tree, "/path/abcde", false, "4")
	assertFound(t, tree, "/pathto/123", false, 5)
}

func TestIsSameSlice(t *testing.T) {
	assert.True(t, isSameSlice([]string{"123", "321", "567"}, []string{"123", "321", "567"}))
	assert.False(t, isSameSlice([]string{"123", "321", "567"}, []string{"321", "123", "567"}))
	assert.False(t, isSameSlice([]string{"321"}, []string{"321", "123"}))
	assert.False(t, isSameSlice([]string{"321", "567"}, []string{"321", "123"}))
}

func TestAddNormalStaticPathShouldOK(t *testing.T) {
	tree := NewTree()
	assertAddExceptFullPath(t, tree, "/", 1, "/")
	assertFound(t, tree, "/", false, 1)

	assertAddExceptFullPath(t, tree, "/abc", 2, "/abc")
	assertFound(t, tree, "/abc", false, 2)
	assertNotFound(t, tree, "/abc/", false)
	assertFound(t, tree, "/abc", true, 2)
	assertFound(t, tree, "/abc/", true, 2)

	assertAddExceptFullPath(t, tree, "/bcde", 3, "/bcde")
	assertFound(t, tree, "/bcde", false, 3)
	assertNotFound(t, tree, "/bcde/", false)
	assertFound(t, tree, "/bcde", true, 3)
	assertFound(t, tree, "/bcde/", true, 3)

	assertAddExceptFullPath(t, tree, "/edef", 4, "/edef")
	assertFound(t, tree, "/edef", false, 4)
	assertNotFound(t, tree, "/edef/", false)
	assertFound(t, tree, "/edef", true, 4)
	assertFound(t, tree, "/edef/", true, 4)

	assertAddExceptFullPath(t, tree, "/ab/dde", 5, "/ab/dde")
	assertFound(t, tree, "/ab/dde", false, 5)
	assertFound(t, tree, "/ab/dde", true, 5)
	assertNotFound(t, tree, "/ab/dde/", false)
	assertFound(t, tree, "/ab/dde/", true, 5)

	assertAddExceptFullPath(t, tree, "/ab", 6, "/ab")
	assertFound(t, tree, "/ab/", true, 6)
	assertNotFound(t, tree, "/ab/", false)

	assertAddExceptFullPath(t, tree, "/a/", 7, "/a/")
	assertNotFound(t, tree, "/a", false)
	assertNotFound(t, tree, "/a", true)
	assertFound(t, tree, "/a/", true, 7)
	assertFound(t, tree, "/a/", false, 7)

}

func TestAddRegexpPathShouldOK(t *testing.T) {
	tree := NewTree()
	assertAddExceptFullPath(t, tree, "/{:\\d+}", 1, "/{:\\d+}")
	assertAddExceptFullPath(t, tree, "/{*:\\d+}", 1, "/{:\\d+}")
	assertAddExceptFullPath(t, tree, "/{p:\\d+}/cde", 2, "/{p:\\d+}/cde")
	assertAddExceptFullPath(t, tree, "/123/{p:\\d+}/cde", 3, "/123/{p:\\d+}/cde")
	assertAddExceptFullPath(t, tree, "/456", 4, "/456")
	assertAddExceptFullPath(t, tree, "/{c:[a-z]{1}}", 5, "/{c:[a-z]{1}}")

	assertFoundParams(t, tree, "/123", false, []*Pair{&Pair{"", "123"}}, 1)
	assertFoundParams(t, tree, "/345", true, []*Pair{&Pair{"", "345"}}, 1)
	assertFoundParams(t, tree, "/123/", true, []*Pair{&Pair{"", "123"}}, 1)
	assertNotFound(t, tree, "/123/", false)
	assertNotFound(t, tree, "/12f", false)

	assertFoundParams(t, tree, "/1/cde", false, []*Pair{&Pair{"p", "1"}}, 2)
	assertFoundParams(t, tree, "/1/cde", true, []*Pair{&Pair{"p", "1"}}, 2)
	assertFoundParams(t, tree, "/1/cde/", true, []*Pair{&Pair{"p", "1"}}, 2)
	assertNotFound(t, tree, "/1/cde/", false)

	assertFoundParams(t, tree, "/123/765/cde", false, []*Pair{&Pair{"p", "765"}}, 3)
	assertFoundParams(t, tree, "/123/34/cde/", true, []*Pair{&Pair{"p", "34"}}, 3)

	assertFoundParams(t, tree, "/456", false, []*Pair{}, 4)
	assertFoundParams(t, tree, "/a", false, []*Pair{&Pair{"c", "a"}}, 5)
	assertNotFound(t, tree, "/a/", false)
	assertFoundParams(t, tree, "/a/", true, []*Pair{&Pair{"c", "a"}}, 5)

	assertNotFound(t, tree, "/abc", false)
	assertNotFound(t, tree, "/abc", true)

}

func TestWildPathShouldOK(t *testing.T) {
	tree := NewTree()
	assertAddExceptFullPath(t, tree, "/{}", 1, "/{}")
	assertAddExceptFullPath(t, tree, "/{p}/a", 2, "/{p}/a")
	assertAddExceptFullPath(t, tree, "/{a}/{b}/{c}", 3, "/{a}/{b}/{c}")
	assertAddExceptFullPath(t, tree, "/{a}/{b}/{c}/", 4, "/{a}/{b}/{c}/")

	assertFoundParams(t, tree, "/abc", false, []*Pair{&Pair{"", "abc"}}, 1)
	assertFoundParams(t, tree, "/abc", true, []*Pair{&Pair{"", "abc"}}, 1)
	assertNotFound(t, tree, "/abc/", false)
	assertFoundParams(t, tree, "/abc/", true, []*Pair{&Pair{"", "abc"}}, 1)

	assertFoundParams(t, tree, "/abc/a", false, []*Pair{&Pair{"p", "abc"}}, 2)
	assertFoundParams(t, tree, "/1/2/3", false, []*Pair{&Pair{"a", "1"}, &Pair{"b", "2"}, &Pair{"c", "3"}}, 3)

}

func TestAddSameWildAndRegexpDifferentParamNameShouldPanic(t *testing.T) {
	tree := NewTree()
	assertAddExceptFullPath(t, tree, "/path/{a}/{b:.+}", "1", "/path/{a}/{b:.+}")
	assertAddExceptFullPath(t, tree, "/path/{a}/{b:.+}", "1", "/path/{a}/{b:.+}")
	assertAddPanic(t, tree, "/path/{c}/{d:.+}")
}

func TestAddWildAndRegexpDuplicateParamNameShouldPanic(t *testing.T) {
	tree := NewTree()
	assertAddPanic(t, tree, "/path/{c}/{c:.+}")
}

//
func TestAddWildAndRegexpWithInvalidParamNameShouldPanic(t *testing.T) {
	tree := NewTree()
	assertAddPanic(t, tree, "/path/to/{-abc}")
	assertAddPanic(t, tree, "/path/to/{abc@#$5}")
	assertAddPanic(t, tree, "/path/to/{-abc:[a-z]+}")
	assertAddPanic(t, tree, "/path/to/{abc@#$5:[a-z]+}")
}

func TestAddInvalidWildOrRegexpPathShouldPanic(t *testing.T) {
	tree := NewTree()
	assertAddPanic(t, tree, "/path/to/a{param}/")
	assertAddPanic(t, tree, "/path/to/{param}_/")
	assertAddPanic(t, tree, "/path/to/a{param:[a-z]{10}}/")
	assertAddPanic(t, tree, "/path/to/{param:[a-z]{10}}_/")
	assertAddPanic(t, tree, "/path/to/{param:}/")
}

func TestAddInvalidRegexpPatternShouldPanic(t *testing.T) {
	tree := NewTree()
	assertAddPanic(t, tree, "/path/to/{:[ab}/")
}

func TestHasStaticBranches(t *testing.T) {
	tree := NewTree()
	tree.Add("/", "0")
	assert.False(t, tree.hasStaticBranches())
	tree.Add("/abc", "1")
	assert.True(t, tree.hasStaticBranches())
	tree.Add("/cde", "2")
	assert.True(t, tree.hasStaticBranches())
}

func TestLookupDoubleSlashShouldNotMatchWildOrRegexp(t *testing.T) {
	tree := NewTree()
	tree.Add("/{param:.*}/{id}/abc", "0")
	assertFound(t, tree, "/123/456/abc", false, "0")
	assertFound(t, tree, "/123/456/abc/", true, "0")
	assertNotFound(t, tree, "//456/abc", false)
	assertNotFound(t, tree, "//456/abc", true)
	assertNotFound(t, tree, "/123//abc", false)
	assertNotFound(t, tree, "/123//abc", true)
}

func TestStaticRegexpWild(t *testing.T) {
	tree := NewTree()
	routes := [][]string{
		{"/", "0"},
		{"/abc", "1"},
		{"/{param:[a-z]+}", "2"},
		{"/a123", "3"},
		{"/{param:[a-z]+}/{id}/r", "5"},
		{"/{param:[a-z]+}/{id}/route", "6"},
		{"/path/to", "7"},
		{"/path/to/{}", "8"},
		{"/path/to/{:\\d+}", "9"},
		{"/path/to/1234", "10"},
		{"/path/to/abc", "11"},
	}
	for _, route := range routes {
		tree.Add(route[0], route[1])
	}

	assertFound(t, tree, "/", false, "0")
	assertFound(t, tree, "/abc", false, "1")
	assertFound(t, tree, "/cdefg", false, "2")
	assertFound(t, tree, "/a123", false, "3")
	assertFound(t, tree, "/abcde/123/r", false, "5")
	assertFound(t, tree, "/cde/0123/r", false, "5")
	assertFound(t, tree, "/abcde/123/route", false, "6")
	assertFound(t, tree, "/cde/0123/route", false, "6")
	assertFound(t, tree, "/path/to", false, "7")
	assertNotFound(t, tree, "/path/to/", false)
	assertFound(t, tree, "/path/to/route", false, "8")
	assertNotFound(t, tree, "/path/to/route/", false)
	assertFound(t, tree, "/path/to/route/", true, "8")
	assertFound(t, tree, "/path/to/123", false, "9")
	assertNotFound(t, tree, "/path/to/123/", false)
	assertFound(t, tree, "/path/to/123/", true, "9")
	assertFound(t, tree, "/path/to/1234", false, "10")
	assertNotFound(t, tree, "/path/to/1234/", false)
	assertFound(t, tree, "/path/to/1234/", true, "10")
	assertFound(t, tree, "/path/to/abc", false, "11")
	assertNotFound(t, tree, "/path/to/abc/", false)
	assertFound(t, tree, "/path/to/abc/", true, "11")
}

func TestAddDoubleSlashesShouldCleanToOneSlash(t *testing.T) {
	tree := NewTree()
	assertAddExceptFullPath(t, tree, "/abc//", 1, "/abc/")
	assertAddExceptFullPath(t, tree, "//cde//", 2, "/cde/")
	assertAddExceptFullPath(t, tree, "/{a}//path", 3, "/{a}/path")
	assertAddExceptFullPath(t, tree, "/{a}//path//{b:.*}", 4, "/{a}/path/{b:.*}")

	assertFound(t, tree, "/abc/", false, 1)
	assertNotFound(t, tree, "/abc//", false)

	assertFound(t, tree, "/cde/", false, 2)
	assertNotFound(t, tree, "//cde//", false)
	assertNotFound(t, tree, "/cde//", false)

	assertFound(t, tree, "/b/path", false, 3)
	assertNotFound(t, tree, "/b//path", false)

	assertFound(t, tree, "/1234/path/4444", false, 4)
	assertNotFound(t, tree, "/1234//path//4444", false)
}
