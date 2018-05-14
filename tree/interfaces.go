package tree

type Pair struct {
	Name  string
	Value string
}

type NodeInterface interface {
	FullPathPattern() string
	Context() interface{}
	Params() []string
}

type AddHookFunc func(context interface{}) interface{}

type TrieInterface interface {
	Lookup(path string, fixTailingSlash bool) (interface{}, []*Pair, bool)
	Add(pattern string, ctx interface{}) NodeInterface
	AddThen(pattern string, callback AddHookFunc) NodeInterface
}
