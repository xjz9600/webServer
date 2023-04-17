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
	var params = map[string]string{}
	segs := strings.Split(path[1:], "/")
	for _, s := range segs {
		child, parents, isParams, found := root.childOf(s)
		if !found {
			if parents.nodeType == STARPATH || parents.nodeType == REPATH {
				root = parents
				break
			}
			return nil, false
		}
		if isParams {
			params[child.path[1:]] = s
		}
		root = child
	}
	if len(params) != 0 {
		return &matchInfo{n: root, params: params}, true
	}
	return &matchInfo{n: root}, true
}

func (r *router) AddRoute(method string, path string, handleFunc HandleFunc) {
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
}

func (n *node) childOrCreate(s string) *node {
	if s[0] == ':' {
		if n.starChild != nil {
			panic("web: 不允许同时注册路径参数和通配符匹配跟正则路由,已有通配符匹配")
		}
		if n.reChild != nil {
			panic("web: 不允许同时注册路径参数和通配符匹配跟正则路由,已有正则匹配")
		}
		n.pathChild = &node{
			path:     s,
			nodeType: PARAMPATH,
		}
		return n.pathChild
	}
	if s == "*" {
		if n.pathChild != nil {
			panic("web: 不允许同时注册路径参数和通配符匹配跟正则路由,已有参数匹配")
		}
		if n.reChild != nil {
			panic("web: 不允许同时注册路径参数和通配符匹配跟正则路由,已有正则匹配")
		}
		n.starChild = &node{
			path:     s,
			nodeType: STARPATH,
		}
		return n.starChild
	}
	if s[0] == '(' && s[len(s)-1] == ')' {
		if n.pathChild != nil {
			panic("web: 不允许同时注册路径参数和通配符匹配跟正则路由,已有参数匹配")
		}
		if n.starChild != nil {
			panic("web: 不允许同时注册路径参数和通配符匹配跟正则路由,已有通配符匹配")
		}
		path := s[1 : len(s)-1]
		reg, err := regexp.Compile(path)
		if err != nil {
			panic(fmt.Sprintf("正则匹配符有问题 (%s)", err.Error()))
		}
		n.reChild = &node{
			path:     path,
			nodeType: REPATH,
			reg:      reg,
		}
		return n.reChild
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

func (n *node) childOf(seg string) (*node, *node, bool, bool) {
	if n.children == nil {
		if n.pathChild != nil {
			return n.pathChild, n, true, true
		}
		if n.reChild != nil {
			if n.reChild.reg.MatchString(seg) {
				return n.reChild, n, false, true
			}
			return nil, n, false, false
		}
		return n.starChild, n, false, n.starChild != nil
	}
	res, ok := n.children[seg]
	if !ok {
		if n.pathChild != nil {
			return n.pathChild, n, true, true
		}
		if n.reChild != nil {
			return n.reChild, n, false, true
		}
		return n.starChild, n, false, n.starChild != nil
	}
	return res, n, false, true
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
}

type matchInfo struct {
	n      *node
	params map[string]string
}
