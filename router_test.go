package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

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
			path:   "/retest/(re.+)",
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
				startChild: &node{
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
								startChild: &node{
									path:    "*",
									handler: mockHandler,
								},
							},
						},
						startChild: &node{
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
						paramChild: &node{
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
							path:    "re.+",
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
	assert.PanicsWithValue(t, "web: 不允许同时注册路径参数和通配符匹配跟正则路由,已有通配符匹配", func() {
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
	if (n.startChild != nil && y.startChild == nil) || (n.startChild == nil && y.startChild != nil) {
		return fmt.Sprintf("通配符节点不匹配"), false
	}
	if n.startChild != nil {
		msg, isOk := n.startChild.equal(y.startChild)
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
	testRoutes := []struct {
		method string
		path   string
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
			path:   "/retest/(re.+)",
		},
	}
	r := NewRouter()
	var mockHandler HandleFunc = func(context *Context) {}
	for _, route := range testRoutes {
		r.AddRoute(route.method, route.path, mockHandler)
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
				startChild: &node{
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
				path:    ":id",
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
				path:    "re.+",
				handler: mockHandler,
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