package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

var mockHandler HandleFunc = func(context *Context) {}

func TestRouter_AddRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/order/create/*",
		},
		{
			method: http.MethodPost,
			path:   "/*",
		},
		{
			method: http.MethodPost,
			path:   "/order/*/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		{
			method: http.MethodPost,
			path:   "/login/:id",
		},
		{
			method: http.MethodTrace,
			path:   "/retest/:id(re.+)",
		},
	}

	var mockHandler HandleFunc = func(context *Context) {}
	r := NewRouter()
	for _, route := range testRoutes {
		r.AddRoute(route.method, route.path, mockHandler)
	}
	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: &node{
				path:    "/",
				handler: mockHandler,
				children: map[string]*node{
					"user": &node{
						path:    "user",
						handler: mockHandler,
						children: map[string]*node{
							"home": &node{
								path:    "home",
								handler: mockHandler,
							},
						},
					},
					"order": &node{
						path: "order",
						children: map[string]*node{
							"detail": &node{
								path:    "detail",
								handler: mockHandler,
							},
						},
					},
				},
			},
			http.MethodPost: &node{
				path: "/",
				starChild: &node{
					path:    "*",
					handler: mockHandler,
				},
				children: map[string]*node{
					"order": &node{
						path: "order",
						children: map[string]*node{
							"create": &node{
								path:    "create",
								handler: mockHandler,
								starChild: &node{
									path:    "*",
									handler: mockHandler,
								},
							},
						},
						starChild: &node{
							path: "*",
							children: map[string]*node{
								"create": &node{
									path:    "create",
									handler: mockHandler,
								},
							},
						},
					},
					"login": &node{
						path:    "login",
						handler: mockHandler,
						pathChild: &node{
							path:    "id",
							handler: mockHandler,
						},
					},
				},
			},
			http.MethodTrace: &node{
				path: "/",
				children: map[string]*node{
					"retest": &node{
						path: "retest",
						reChild: &node{
							path:    "id",
							handler: mockHandler,
						},
					},
				},
			},
		},
	}
	msg, ok := wantRouter.equal(r)
	assert.True(t, ok, msg)

	r = NewRouter()
	assert.PanicsWithValue(t, "web: 路径不能为空字符串", func() {
		r.AddRoute(http.MethodGet, "", mockHandler)
	})
	assert.PanicsWithValue(t, "web: 路径必须以 [/] 开头", func() {
		r.AddRoute(http.MethodGet, "login", mockHandler)
	})
	assert.PanicsWithValue(t, "web: 路径不能以 [/] 结尾", func() {
		r.AddRoute(http.MethodGet, "/login/", mockHandler)
	})
	assert.PanicsWithValue(t, "web: 不能有连续的 //", func() {
		r.AddRoute(http.MethodGet, "/login//ab", mockHandler)
	})
	r.AddRoute(http.MethodGet, "/", mockHandler)
	assert.PanicsWithValue(t, "web: 路由冲突，重复注册[/]", func() {
		r.AddRoute(http.MethodGet, "/", mockHandler)
	})
	r.AddRoute(http.MethodGet, "/abc/ab", mockHandler)
	assert.PanicsWithValue(t, "web: 路由冲突，重复注册[/abc/ab]", func() {
		r.AddRoute(http.MethodGet, "/abc/ab", mockHandler)
	})
	r.AddRoute(http.MethodGet, "/abc/*", mockHandler)
	assert.PanicsWithValue(t, "web: 非法路由，已有通配符路由。不允许同时注册通配符路由和参数路由 [id]", func() {
		r.AddRoute(http.MethodGet, "/abc/:id", mockHandler)
	})
	r.AddRoute(http.MethodGet, "/mmm/:id", mockHandler)
	assert.PanicsWithValue(t, "web: 不允许同时注册路径参数和通配符匹配跟正则路由,已有参数匹配", func() {
		r.AddRoute(http.MethodGet, "/mmm/*", mockHandler)
	})

}

func (r router) equal(y router) (string, bool) {
	for k, v := range r.trees {
		dst, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("找不到对应的 http method (%s)", k), false
		}
		msg, isOk := v.equal(dst)
		if !isOk {
			return msg, isOk
		}
	}
	return "", true
}

func (n *node) equal(y *node) (string, bool) {
	if n.path != y.path {
		return fmt.Sprintf("节点路径不匹配 (%s)", n.path), false
	}
	if len(n.children) != len(y.children) {
		return fmt.Sprintf("子节点路径数量不匹配"), false
	}
	if (n.starChild != nil && y.starChild == nil) || (n.starChild == nil && y.starChild != nil) {
		return fmt.Sprintf("通配符节点不匹配"), false
	}
	if n.starChild != nil {
		msg, isOk := n.starChild.equal(y.starChild)
		if !isOk {
			return msg, isOk
		}
	}
	hHandler := reflect.ValueOf(n.handler)
	yHandler := reflect.ValueOf(y.handler)
	if hHandler != yHandler {
		return fmt.Sprintf("handler 不相等"), false
	}
	for p, c := range n.children {
		dst, ok := y.children[p]
		if !ok {
			return fmt.Sprintf("子节点路径不存在 (%s)", p), false
		}
		msg, isOk := c.equal(dst)
		if !isOk {
			return msg, isOk
		}
	}
	return "", true
}

func TestRouter_FindRoute(t *testing.T) {
	var firstHandler HandleFunc = func(context *Context) {
		fmt.Println("aa")
	}
	var sencodHandler HandleFunc = func(context *Context) {
		fmt.Println("bb")
	}
	var thirdHandler HandleFunc = func(context *Context) {
		fmt.Println("cc")
	}
	var fourHandler HandleFunc = func(context *Context) {
		fmt.Println("dd")
	}
	testRoutes := []struct {
		method  string
		path    string
		handler HandleFunc
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodPut,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		{
			method: http.MethodPut,
			path:   "/*",
		},
		{
			method: http.MethodGet,
			path:   "/*",
		},
		{
			method: http.MethodPost,
			path:   "/order/*/create",
		},
		{
			method: http.MethodPost,
			path:   "/myparams/:id",
		},
		{
			method: http.MethodPost,
			path:   "/testParams/*",
		},
		{
			method: http.MethodTrace,
			path:   "/retest/:id(re.+)",
		},
		{
			method:  http.MethodGet,
			path:    "/a/b/c/d/*",
			handler: firstHandler,
		},
		{
			method:  http.MethodGet,
			path:    "/*/b/c/d/e",
			handler: sencodHandler,
		},
		{
			method:  http.MethodGet,
			path:    "/a/*/c/*/e",
			handler: thirdHandler,
		},
		{
			method:  http.MethodGet,
			path:    "/a/b/c/d/e",
			handler: fourHandler,
		},
	}
	r := NewRouter()
	for _, route := range testRoutes {
		if route.handler != nil {
			r.AddRoute(route.method, route.path, route.handler)
		} else {
			r.AddRoute(route.method, route.path, mockHandler)
		}

	}

	testCases := []struct {
		name      string
		method    string
		path      string
		wantNode  *node
		wantFound bool
		params    map[string]string
	}{
		{
			name:      "没有路由",
			method:    http.MethodDelete,
			path:      "/",
			wantFound: false,
		},
		{
			name:      "order detail",
			method:    http.MethodGet,
			path:      "/order/detail",
			wantFound: true,
			wantNode: &node{
				path:    "detail",
				handler: mockHandler,
			},
		},
		{
			name:      "root",
			method:    http.MethodPut,
			path:      "/",
			wantFound: true,
			wantNode: &node{
				path:    "/",
				handler: mockHandler,
				starChild: &node{
					path:    "*",
					handler: mockHandler,
				},
			},
		},
		{
			name:      "star root",
			method:    http.MethodPut,
			path:      "/abc",
			wantFound: true,
			wantNode: &node{
				path:    "*",
				handler: mockHandler,
			},
		},
		{
			name:      "star order",
			method:    http.MethodPost,
			path:      "/order/adsf/create",
			wantFound: true,
			wantNode: &node{
				path:    "create",
				handler: mockHandler,
			},
		},
		{
			name:      "params myparams",
			method:    http.MethodPost,
			path:      "/myparams/123",
			wantFound: true,
			wantNode: &node{
				path:    "id",
				handler: mockHandler,
			},
			params: map[string]string{
				"id": "123",
			},
		},
		{
			name:      "star multipath",
			method:    http.MethodPost,
			path:      "/testParams/f1/f2/f3",
			wantFound: true,
			wantNode: &node{
				path:    "*",
				handler: mockHandler,
			},
		},
		{
			name:      "rz retest",
			method:    http.MethodTrace,
			path:      "/retest/reMyTest",
			wantFound: true,
			wantNode: &node{
				path:    "id",
				handler: mockHandler,
			},
			params: map[string]string{
				"id": "reMyTest",
			},
		},
		{
			name:      "start /a/b/c/d/e",
			method:    http.MethodGet,
			path:      "/a/b/c/d/e",
			wantFound: true,
			wantNode: &node{
				path:    "e",
				handler: fourHandler,
			},
		},
		{
			name:      "start /a/b/c/d/m",
			method:    http.MethodGet,
			path:      "/a/b/c/d/m",
			wantFound: true,
			wantNode: &node{
				path:    "*",
				handler: firstHandler,
			},
		},
		{
			name:      "start /c/b/c/d/e",
			method:    http.MethodGet,
			path:      "/c/b/c/d/e",
			wantFound: true,
			wantNode: &node{
				path:    "e",
				handler: sencodHandler,
			},
		},
		{
			name:      "start /a/m/c/m/e",
			method:    http.MethodGet,
			path:      "/a/m/c/m/e",
			wantFound: true,
			wantNode: &node{
				path:    "e",
				handler: thirdHandler,
			},
		},
		{
			name:      "start /a/b/c",
			method:    http.MethodGet,
			path:      "/a/b/c",
			wantFound: true,
			wantNode: &node{
				path:    "*",
				handler: mockHandler,
				children: map[string]*node{
					"b": &node{
						path: "b",
						children: map[string]*node{
							"c": &node{
								path: "c",
								children: map[string]*node{
									"d": &node{
										path: "d",
										children: map[string]*node{
											"e": &node{
												path:    "e",
												handler: sencodHandler,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info, found := r.FindRoute(tc.method, tc.path)
			assert.Equal(t, tc.wantFound, found)
			if !found {
				return
			}
			msg, isOK := info.n.equal(tc.wantNode)
			assert.Equal(t, tc.params, info.params)
			assert.True(t, isOK, msg)
		})
	}
}

func TestRouter_FindRoute_Middleware(t *testing.T) {
	var mdsBuilder = func(i byte) Middleware {
		return func(next HandleFunc) HandleFunc {
			return func(context *Context) {
				context.RespData = append(context.RespData, i)
				next(context)
			}
		}
	}
	mdsRouter := []struct {
		method string
		path   string
		mds    []Middleware
	}{
		{
			method: http.MethodGet,
			path:   "/a/*",
			mds:    []Middleware{mdsBuilder('a'), mdsBuilder('*')},
		},
		{
			method: http.MethodGet,
			path:   "/a/b/c",
			mds:    []Middleware{mdsBuilder('a'), mdsBuilder('b'), mdsBuilder('c')},
		},
		{
			method: http.MethodGet,
			path:   "/a/*/c",
			mds:    []Middleware{mdsBuilder('a'), mdsBuilder('*'), mdsBuilder('d')},
		},
		{
			method: http.MethodGet,
			path:   "/a/b/:id",
			mds:    []Middleware{mdsBuilder('a'), mdsBuilder(':'), mdsBuilder('d')},
		},
	}
	r := NewRouter()
	for _, md := range mdsRouter {
		r.AddRoute(md.method, md.path, mockHandler, md.mds...)
	}
	testCases := []struct {
		name     string
		method   string
		path     string
		wantResp string
	}{
		{
			name:     "star middleware",
			method:   http.MethodGet,
			path:     "/a/b/m",
			wantResp: "a*a:d",
		},
		{
			name:     "star middleware",
			method:   http.MethodGet,
			path:     "/a/m/d",
			wantResp: "a*",
		},
		{
			name:     "star middleware",
			method:   http.MethodGet,
			path:     "/a/b/c",
			wantResp: "a*abca:da*d",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mi, _ := r.FindRoute(tc.method, tc.path)
			mdls := mi.mds
			var root HandleFunc = func(ctx *Context) {
				// 使用 string 可读性比较高
				assert.Equal(t, tc.wantResp, string(ctx.RespData))
			}
			for i := len(mdls) - 1; i >= 0; i-- {
				root = mdls[i](root)
			}
			// 开始调度
			root(&Context{
				RespData: make([]byte, 0, len(tc.wantResp)),
			})
		})
	}
}
