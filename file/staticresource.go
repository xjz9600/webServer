package file

import (
	"github.com/hashicorp/golang-lru/v2"
	"net/http"
	"os"
	"path/filepath"
	web "routing"
	"strconv"
)

type staticResourceBuilder struct {
	dir                     string
	cache                   *lru.Cache[string, any]
	extensionContentTypeMap map[string]string
	maxSize                 int
}

type staticResourceBuilderOptions func(*staticResourceBuilder)

func NewstaticResourceBuilder(opts ...staticResourceBuilderOptions) (*staticResourceBuilder, error) {
	c, err := lru.New[string, any](1000)
	if err != nil {
		return nil, err
	}
	res := &staticResourceBuilder{
		dir:   filepath.Join("testdata", "static"),
		cache: c,
		extensionContentTypeMap: map[string]string{
			// 这里根据自己的需要不断添加
			"jpeg": "image/jpeg",
			"jpe":  "image/jpeg",
			"jpg":  "image/jpeg",
			"png":  "image/png",
			"pdf":  "application/pdf",
		},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func StaticWithMaxFileSize(maxSize int) staticResourceBuilderOptions {
	return func(handler *staticResourceBuilder) {
		handler.maxSize = maxSize
	}
}

func StaticWithDir(dir string) staticResourceBuilderOptions {
	return func(handler *staticResourceBuilder) {
		handler.dir = dir
	}
}

func StaticWithCache(c *lru.Cache[string, any]) staticResourceBuilderOptions {
	return func(handler *staticResourceBuilder) {
		handler.cache = c
	}
}

func StaticWithMoreExtension(extMap map[string]string) staticResourceBuilderOptions {
	return func(h *staticResourceBuilder) {
		for ext, contentType := range extMap {
			h.extensionContentTypeMap[ext] = contentType
		}
	}
}

func (s *staticResourceBuilder) Handle(ctx *web.Context) {
	fileName, err := ctx.PathValue("file").AsString()
	if err != nil {
		ctx.RespData = []byte("请传入文件路径")
		ctx.RespStatusCode = http.StatusInternalServerError
		return
	}
	dst := filepath.Join(s.dir, fileName)
	ext := filepath.Ext(dst)
	if len(ext) == 0 {
		ctx.RespData = []byte("请传入正确文件名")
		ctx.RespStatusCode = http.StatusInternalServerError
		return
	}
	ext = ext[1:]
	header := ctx.Resp.Header()
	if data, ok := s.cache.Get(dst); ok {
		contentType := s.extensionContentTypeMap[ext]
		header.Set("Content-Type", contentType)
		header.Set("Content-Length", strconv.Itoa(len(data.([]byte))))
		ctx.RespData = data.([]byte)
		ctx.RespStatusCode = 200
		return
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		ctx.RespStatusCode = http.StatusInternalServerError
		ctx.RespData = []byte("文件未找到")
		return
	}
	// 大文件不缓存
	if len(data) <= s.maxSize {
		s.cache.Add(dst, data)
	}
	// 可能的有文本文件，图片，多媒体（视频，音频）
	contentType := s.extensionContentTypeMap[ext]
	header.Set("Content-Type", contentType)
	header.Set("Content-Length", strconv.Itoa(len(data)))
	ctx.RespData = data
	ctx.RespStatusCode = 200
}
