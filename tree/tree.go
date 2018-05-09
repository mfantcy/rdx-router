// Radix tree mix regexp and wild implementation
//
// Format
// A route pattern placeholder starts with a {, followed by the placeholder name, ending with a }.
// This is an example placeholder named name:
// "/hello/{name}"
// No param
// "/hello/{} | /hello/{*}"
// Regular expression matching
// "/users/{id:[0-9]+}"
// No param
// "/numbers/{:[0-9]+}|/numbers/{*:[0-9]+}"

package tree

import (
	"errors"
	"regexp"
)

type nodeType uint8

const (
	nodeTypeStatic nodeType = iota
	nodeTypeRegexp
	nodeTypeWild
)

type Pair struct {
	Key   string
	Value string
}

type Pairs []*Pair

type Edge interface {
	FullPathPattern() string
	Context() interface{}
	Params() []string
}

type AddHookFunc func(context interface{}) interface{}

type Tree interface {
	Lookup(path string, fixTailingSlash bool) (interface{}, Pairs, bool)
	Add(pattern string, ctx interface{}) Edge
	AddThen(pattern string, callback AddHookFunc) Edge
}

type leaf struct {
	params  []string
	context interface{}
}

type node struct {
	path           string
	pathLen        int
	nodeType       nodeType
	parent         *node
	hasNonStatic   bool
	staticBranches []*node
	regexpBranches []*node
	wildBranch     *node
	regexp         *regexp.Regexp
	leaf           *leaf
}

func (n *node) FullPathPattern() string {
	var params []string
	if n.leaf != nil {
		params = n.leaf.params
	}
	cn := n
	path := ""
	pIdx := len(params) - 1
	for cn != nil {
		p := cn.path
		if cn.nodeType == nodeTypeWild {
			p = "{}"
			if pIdx >= 0 {
				p = "{" + params[pIdx] + "}"
				pIdx--
			}
		} else if cn.nodeType == nodeTypeRegexp {
			p = "{:" + cn.path + "}"
			if pIdx >= 0 {
				p = "{" + params[pIdx] + ":" + cn.path + "}"
				pIdx--
			}
		}
		path = p + path
		cn = cn.parent
	}
	return path
}

func (n *node) Context() (ctx interface{}) {
	if n.leaf != nil {
		ctx = n.leaf.context
	}
	return
}

func (n *node) Params() (params []string) {
	if n.leaf != nil {
		params = n.leaf.params
	}
	return
}

func (n *node) Add(pattern string, ctx interface{}) Edge {
	return n.AddThen(pattern, func(context interface{}) interface{} {
		return ctx
	})
}

func (n *node) AddThen(pattern string, callback AddHookFunc) Edge {
	var params []string
	treetop := n.add(pattern, &params)
	if treetop.leaf != nil {
		if !isSameSlice(treetop.leaf.params, params) {
			panic(errors.New("param in path is conflict prev registered"))
		}
	} else {
		treetop.leaf = &leaf{params: params}
	}
	if callback != nil && treetop.leaf != nil {
		treetop.leaf.context = callback(treetop.leaf.context)
	}
	return treetop
}

func NewTree() *node {
	return newNode()
}

func newNode() *node {
	return &node{staticBranches: make([]*node, 256)}
}

func isSameSlice(a []string, b []string) bool {
	if len(a) == len(b) {
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}
	} else {
		return false
	}
	return true
}

func (n *node) Lookup(path string, fixTailingSlash bool) (ctx interface{}, pairs Pairs, ok bool) {
	var leaf *leaf
	if leaf, pairs = n.lookUp(path, fixTailingSlash); leaf != nil {
		return leaf.context, pairs, true
	}
	return ctx, pairs, false
}

func paramsAppend(params []string, param string) []string {
	if param == "*" {
		param = ""
	}
	if param != "" {
		for _, v := range params {
			if v == param {
				panic(errors.New("param name '" + param + "' duplicate"))
			}
		}
	}
	return append(params, param)
}

func (n *node) add(path string, params *[]string) *node {
	nodeType, param, regexPattern, bracesPos, bracesLen := determinePlaceholder(path)
	if nodeType == nodeTypeWild || nodeType == nodeTypeRegexp {
		*params = paramsAppend(*params, param)
		if nodeType == nodeTypeWild {
			return n.insertStaticNode(path[:bracesPos]).
				insertWildNode(path[bracesPos+bracesLen:], params)
		} else {
			return n.insertStaticNode(path[:bracesPos]).
				insertRegexpNode(regexPattern, path[bracesPos+bracesLen:], params)
		}
	}
	return n.insertStaticNode(path)
}

func determinePlaceholder(str string) (nodeType nodeType, param string, regexpPattern string, bracesPos int, bracesLen int) {
	i, bracesStack, bracesEnd := 0, 0, 0
	backSlashOpen := false
	regexpPattern = ""
	for ; i < len(str); i++ {
		if bracesPos == 0 { //lookUp for placeholder start "/"
			if str[i] == '{' {
				if i == 0 || str[i-1] != '/' {
					panic(errors.New("placeholder \"{\" must be followed by \"/\""))
				}
				bracesPos = i
			}
		} else if nodeType == nodeTypeRegexp {
			if str[i] == '{' && !backSlashOpen { //stack regexp braces
				bracesStack++
			} else if str[i] == '}' && !backSlashOpen {
				if bracesStack == 0 {
					if len(regexpPattern) == 0 {
						panic(errors.New("regexp pattern is empty"))
					}
					bracesEnd = i
					break
				} else {
					bracesStack--
				}
			} else if backSlashOpen {
				backSlashOpen = false
			} else if str[i] == '\\' {
				backSlashOpen = true
			}
			regexpPattern += string(str[i])
		} else if bracesPos > 0 && bracesEnd == 0 {
			if str[i] == ':' {
				nodeType = nodeTypeRegexp
			} else if str[i] == '}' {
				bracesEnd = i
				break
			} else {
				param += string(str[i])
			}
		}
	}
	if bracesPos > 0 && str[bracesEnd] == '}' {
		if i+1 != len(str) && str[i+1] != '/' {
			panic(errors.New("placeholder \"}\" must be at the end or before \"/\""))
		}
		if nodeType != nodeTypeRegexp {
			nodeType = nodeTypeWild
		}
		reg, _ := regexp.Compile("^$|^\\*$|^[a-zA-Z0-9_]+(-*[a-zA-Z0-9_]+)*$")
		if !reg.Match([]byte(param)) {
			panic(errors.New("invalid param name \"" + param + "\""))
		}
		bracesLen = (bracesEnd - bracesPos) + 1
	} else {
		nodeType, param, regexpPattern, bracesPos, bracesLen = nodeTypeStatic, "", "", 0, 0
	}
	return
}

func (n *node) insertStaticNode(path string) *node {
	path = cleanDuplicateSlash(path)
	//new static node
	if len(n.path) == 0 && !n.hasStaticBranches() {
		n.path = path
		n.pathLen = len(n.path)
		n.nodeType = nodeTypeStatic
		return n
	}
	i := 0
	if n.nodeType == nodeTypeStatic {
		max := min(len(path), len(n.path))
		for i < max && path[i] == n.path[i] {
			i++
		}
		if i < len(n.path) {
			n.splitEdge(i)
		}
	}

	if i < len(path) {
		firstByte := path[i:][0]
		b := n.staticBranches[firstByte]
		if b == nil {
			b = newNode()
			b.parent = n
			n.staticBranches[firstByte] = b
		}
		return b.insertStaticNode(path[i:])
	}
	return n
}

func (n *node) insertWildNode(tail string, params *[]string) *node {
	n.hasNonStatic = true
	if n.wildBranch == nil {
		n.wildBranch = newNode()
		n.wildBranch.parent = n
		n.wildBranch.nodeType = nodeTypeWild
		n.wildBranch.path = "{}"
	}
	tail = clearPrefixSlash(tail)
	if tail == "" {
		return n.wildBranch
	}
	return n.wildBranch.add(tail, params)
}

func (n *node) insertRegexpNode(regexPattern string, tail string, params *[]string) *node {
	var regexpNode *node
	n.hasNonStatic = true
	for k := range n.regexpBranches {
		if regexPattern == n.regexpBranches[k].regexp.String() {
			regexpNode = n.regexpBranches[k]
			break
		}
	}
	if regexpNode == nil {
		regexpNode = newNode()
		regexpNode.path = regexPattern
		regexpNode.nodeType = nodeTypeRegexp
		regexpNode.parent = n
		if rx, err := regexp.Compile(regexPattern); err != nil {
			panic(err)
		} else {
			regexpNode.regexp = rx
		}
		n.regexpBranches = append(n.regexpBranches, regexpNode)
	}

	tail = clearPrefixSlash(tail)
	if tail == "" {
		return regexpNode
	}
	return regexpNode.add(tail, params)
}

func clearPrefixSlash(p string) (ret string) {
	rxp, _ := regexp.Compile("^/+")
	ret = string(rxp.ReplaceAll([]byte(p), []byte{'/'}))
	return
}

func cleanDuplicateSlash(p string) (ret string) {
	rxp, _ := regexp.Compile("/+")
	ret = string(rxp.ReplaceAll([]byte(p), []byte{'/'}))
	return
}

func (n *node) hasStaticBranches() bool {
	for _, child := range n.staticBranches {
		if child != nil {
			return true
		}
	}
	return false
}

func (n *node) updateParentOfBranches() {
	for _, c := range n.staticBranches {
		if c != nil {
			c.parent = n
		}
	}
	for _, c := range n.regexpBranches {
		if c != nil {
			c.parent = n
		}
	}
	if n.wildBranch != nil {
		n.wildBranch.parent = n
	}
}

func (n *node) splitEdge(pos int) {
	branch := &node{
		path:           n.path[pos:],
		pathLen:        len(n.path[pos:]),
		nodeType:       n.nodeType,
		parent:         n,
		hasNonStatic:   n.hasNonStatic,
		staticBranches: n.staticBranches,
		regexpBranches: n.regexpBranches,
		wildBranch:     n.wildBranch,
		regexp:         n.regexp,
		leaf:           n.leaf,
	}
	branch.updateParentOfBranches()
	n.path = n.path[:pos]
	n.pathLen = len(n.path)
	n.staticBranches = make([]*node, 256)
	n.staticBranches[branch.path[0]] = branch
	n.regexpBranches = n.regexpBranches[:0]
	n.wildBranch = nil
	n.hasNonStatic = false
	n.regexp = nil
	n.leaf = nil
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

type nodeStack struct {
	prev      *nodeStack
	node      *node
	po        int
	regexpIdx int
	isWildCmp bool
	cmps      int
	cmpl      int
}

type backStateStack struct {
	//previous backBtack
	prev *backStateStack
	//state node
	node *node
	//path start position
	po int
	//is prefixMatch
	prefixMatch bool
	//regexp index
	regexpIdx int
	//param position start
	paramPo int
	//param length
	paramLen int
	//wild node process done
	wildDone bool
}

func (n *node) lookUp(path string, fixTailingSlash bool) (*leaf, Pairs) { //to speed up , no others func call
	var backStack, pevStack *backStateStack
	var leaf *leaf
	var next *node
	var prefixMatch, wildDone bool
	var po, regexpIdx, paramPo, paramLen int

walk:
	next = nil
	nextPo := 0
	switch n.nodeType {
	case nodeTypeStatic:
		if !prefixMatch {
			regexpIdx, paramPo, paramLen, wildDone = 0, 0, 0, false
			if len(path[po:]) > n.pathLen {
				if p := path[po : po+n.pathLen]; p == n.path {
					prefixMatch = true
					if npo := po + n.pathLen; n.staticBranches[path[npo]] != nil {
						nextPo = npo
						next = n.staticBranches[path[npo]]
						goto beforeNext
					} else if fixTailingSlash && path[po+n.pathLen] == '/' && n.leaf != nil {
						leaf = n.leaf
						goto found
					}
				}
			} else if path[po:] == n.path && n.leaf != nil {
				leaf = n.leaf
				goto found

			} else if fixTailingSlash && path[po:] == "/" && n.path[0] == '/' {
				if n.parent != nil && n.parent.leaf != nil {
					leaf = n.parent.leaf
					goto found
				}
			}
		}
		if prefixMatch {
			if paramLen == 0 {
				paramPo = po + n.pathLen
				for paramLen < len(path[paramPo:]) && path[paramPo+paramLen] != '/' {
					paramLen++
				}
			}
			//double slash
			if paramLen == 0 {
				break
			}
			nextPo = paramPo
			//regexp comparison
			for regexpIdx < len(n.regexpBranches) {
				regexpNode := n.regexpBranches[regexpIdx]
				regexpIdx++

				if m := regexpNode.regexp.FindAllString(path[paramPo:paramPo+paramLen], 1); len(m) > 0 && m[0] == path[paramPo:paramPo+paramLen] {
					next = regexpNode
					goto beforeNext
				}
			}

			//wild
			if !wildDone {
				wildDone = true
				if n.wildBranch != nil {
					next = n.wildBranch
					goto beforeNext
				}
			}
		}
	case nodeTypeWild, nodeTypeRegexp:
		if npo := po + paramLen; len(path) == npo && n.leaf != nil {
			leaf = n.leaf
			goto found
		} else if len(path) > npo {
			if n.staticBranches[path[npo]] != nil {
				po = npo
				n = n.staticBranches[path[npo]]
				goto walk
			} else if fixTailingSlash && path[npo:] == "/" && n.leaf != nil {
				leaf = n.leaf
				goto found
			}
		}

	}
	if backStack != nil {
		n = backStack.node
		po = backStack.po
		prefixMatch = backStack.prefixMatch
		wildDone = backStack.wildDone
		regexpIdx = backStack.regexpIdx
		paramPo = backStack.paramPo
		paramLen = backStack.paramLen
		pevStack = backStack
		backStack = backStack.prev
		goto walk
	}
	return nil, Pairs{}

beforeNext:
	if next != nil {
		if n.hasNonStatic {
			if backStack == nil || backStack.node != n {
				if pevStack != nil && pevStack.node == n {
					backStack = pevStack
				} else {
					backStack = &backStateStack{
						prev:        backStack,
						node:        n,
						po:          po,
						prefixMatch: prefixMatch,
					}
				}
			}
			backStack.regexpIdx = regexpIdx
			backStack.paramPo = paramPo
			backStack.paramLen = paramLen
			backStack.wildDone = wildDone
		}
		n = next
		po = nextPo
		prefixMatch = false
		goto walk
	}
found:
	params := make(Pairs, len(leaf.params), len(leaf.params))
	paramsIdx := len(params) - 1
	for paramsIdx >= 0 && backStack != nil {
		if backStack.paramLen > 0 {
			params[paramsIdx] = &Pair{leaf.params[paramsIdx], path[backStack.paramPo : backStack.paramPo+backStack.paramLen]}
			paramsIdx--
		}
		backStack = backStack.prev
	}
	return leaf, params
}
