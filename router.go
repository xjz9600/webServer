package web

import (
	"fmt"
	"regexp"
	"strings"
)

type NodeType int

const (
	FULLPATH NodeType = iota
	PARAMPATH
	STARPATH
	REPATH
)

type router struct {
	trees map[string]*node
}

func (r *router) FindRoute(method string, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}
	if path == "/" {
		return &matchInfo{n: root}, true
	}
	segs := strings.Split(path[1:], "/")
	cur, params, found, mds := r.findNodeAndMds(root, segs)
	if !found {
		return nil, false
	}
	res := &matchInfo{
		n: cur,
	}
	if len(mds) > 0 {
		res.mds = mds
	}
	if len(params) > 0 {
		res.params = params
	}
	return res, true
}

func (r *router) findNodeAndMds(root *node, segs []string) (*node, map[string]string, bool, []Middleware) {
	queue := []*node{root}
	var resNode *node
	var mds []Middleware
	var params map[string]string
	if len(root.mds) > 0 {
		mds = append(mds, root.mds...)
	}
	for _, s := range segs {
		var cur []*node
		for _, q := range queue {
			children, childrenMds := q.findNodeChildren(s)
			if q.handler != nil && q.nodeType == STARPATH {
				resNode = q
			}
			if q.nodeType == PARAMPATH || q.nodeType == REPATH {
				if params == nil {
					params = make(map[string]string)
				}
				params[q.path] = s
			}
			cur = append(cur, children...)
			mds = append(mds, childrenMds...)
		}
		queue = cur
	}
	if len(queue) > 0 {
		for i := 0; i < len(queue); i++ {
			if queue[i].nodeType == PARAMPATH || queue[i].nodeType == REPATH {
				if params == nil {
					params = make(map[string]string)
				}
				params[queue[i].path] = segs[len(segs)-1]
			}
			if queue[i].handler != nil {
				return queue[i], params, true, mds
			}
		}
	}
	if resNode != nil {
		return resNode, params, true, mds
	}
	return nil, nil, false, mds
}

func (n *node) findNodeMds() []Middleware {
	if len(n.mds) > 0 {
		return n.mds
	}
	return []Middleware{}
}
func (n *node) findNodeChildren(s string) ([]*node, []Middleware) {
	var res []*node
	var mds []Middleware
	if n.children != nil {
		if re, ok := n.children[s]; ok {
			res = append(res, re)
			mds = append(mds, re.findNodeMds()...)
		}
	}
	if n.reChild != nil && n.reChild.reg.MatchString(s) {
		res = append(res, n.reChild)
		mds = append(mds, n.reChild.findNodeMds()...)
	}
	if n.pathChild != nil {
		res = append(res, n.pathChild)
		mds = append(mds, n.pathChild.findNodeMds()...)
	}
	if n.starChild != nil {
		res = append(res, n.starChild)
		mds = append(mds, n.starChild.findNodeMds()...)
	}
	return res, mds
}

func (n *node) parseParam(path string) (string, string, bool) {
	// 去除 :
	path = path[1:]
	// paramName xxx
	segs := strings.SplitN(path, "(", 2)
	if len(segs) == 2 {
		expr := segs[1]
		if strings.HasSuffix(expr, ")") {
			return segs[0], expr[:len(expr)-1], true
		}
	}
	return path, "", false
}
func (n *node) childOrCreateReg(path string, expr string) *node {
	if n.starChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [%s]", path))
	}
	if n.pathChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册正则路由和参数路由 [%s]", path))
	}
	if n.reChild != nil {
		if n.reChild.reg.String() != expr || n.path != path {
			panic(fmt.Sprintf("web: 路由冲突，正则路由冲突，已有 %s，新注册 %s", n.reChild.path, path))
		}
	} else {
		regExpr, err := regexp.Compile(expr)
		if err != nil {
			panic(fmt.Errorf("web: 正则表达式错误 %w", err))
		}
		n.reChild = &node{path: path, reg: regExpr, nodeType: REPATH}
	}
	return n.reChild
}

func (n *node) childOrCreateParam(path string) *node {
	if n.reChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有正则路由。不允许同时注册正则路由和参数路由 [%s]", path))
	}
	if n.starChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [%s]", path))
	}
	if n.pathChild != nil {
		if n.pathChild.path != path {
			panic(fmt.Sprintf("web: 路由冲突，参数路由冲突，已有 %s，新注册 %s", n.pathChild.path, path))
		}
	} else {
		n.pathChild = &node{path: path, nodeType: PARAMPATH}
	}
	return n.pathChild
}

func (n *node) childOrCreate(s string) *node {
	if s[0] == ':' {
		path, expr, isReg := n.parseParam(s)
		if isReg {
			return n.childOrCreateReg(path, expr)
		}

		return n.childOrCreateParam(path)
	}
	if s == "*" {
		if n.pathChild != nil {
			panic("web: 不允许同时注册路径参数和通配符匹配跟正则路由,已有参数匹配")
		}
		if n.reChild != nil {
			panic("web: 不允许同时注册路径参数和通配符匹配跟正则路由,已有正则匹配")
		}
		if n.starChild == nil {
			n.starChild = &node{
				path:     s,
				nodeType: STARPATH,
			}
		}
		return n.starChild
	}
	if n.children == nil {
		n.children = map[string]*node{}
	}
	res, ok := n.children[s]
	if !ok {
		res = &node{
			path:     s,
			nodeType: FULLPATH,
		}
		n.children[s] = res
	}
	return res
}

func (r *router) AddRoute(method string, path string, handleFunc HandleFunc, mds ...Middleware) {
	if path == "" {
		panic("web: 路径不能为空字符串")
	}
	root, ok := r.trees[method]
	if !ok {
		root = &node{
			path:     "/",
			nodeType: FULLPATH,
		}
		r.trees[method] = root
	}
	if path == "/" {
		if root.handler != nil {
			panic("web: 路由冲突，重复注册[/]")
		}
		root.handler = handleFunc
		root.route = "/"
		root.mds = mds
		return
	}
	if path[0] != '/' {
		panic("web: 路径必须以 [/] 开头")
	}
	if path[len(path)-1] == '/' {
		panic("web: 路径不能以 [/] 结尾")
	}
	segs := strings.Split(path[1:], "/")
	for _, s := range segs {
		if s == "" {
			panic("web: 不能有连续的 //")
		}
		child := root.childOrCreate(s)
		root = child
	}
	if root.handler != nil {
		panic(fmt.Sprintf("web: 路由冲突，重复注册[%s]", path))
	}
	root.handler = handleFunc
	root.route = path
	root.mds = mds
}

func (n *node) childOf(seg string) (*node, bool, bool) {
	if n.children == nil {
		if n.pathChild != nil {
			return n.pathChild, true, true
		}
		if n.reChild != nil {
			if n.reChild.reg.MatchString(seg) {
				return n.reChild, false, true
			}
			return nil, false, false
		}
		return n.starChild, false, n.starChild != nil
	}
	res, ok := n.children[seg]
	if !ok {
		if n.pathChild != nil {
			return n.pathChild, true, true
		}
		if n.reChild != nil {
			return n.reChild, false, true
		}
		return n.starChild, false, n.starChild != nil
	}
	return res, false, true
}

func NewRouter() router {
	return router{trees: make(map[string]*node)}
}

type node struct {
	path      string
	children  map[string]*node
	handler   HandleFunc
	starChild *node
	pathChild *node
	nodeType  NodeType
	reChild   *node
	reg       *regexp.Regexp
	mds       []Middleware
	route     string
}

type matchInfo struct {
	n      *node
	mds    []Middleware
	params map[string]string
}
