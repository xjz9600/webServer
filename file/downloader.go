package file

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	web "routing"
	"strings"
)

type downLoader struct {
	dir string
}

type downLoadOptions func(*downLoader)

func NewDownLoader(opts ...downLoadOptions) *downLoader {
	res := &downLoader{
		dir: filepath.Join("testdata", "download"),
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func DownloaderWithDir(dir string) downLoadOptions {
	return func(loader *downLoader) {
		loader.dir = dir
	}
}

func (d *downLoader) Handle(ctx *web.Context) {
	fileName, err := ctx.FormValue("file").AsString()
	if err != nil {
		ctx.RespData = []byte("请输出文件名称")
		ctx.RespStatusCode = http.StatusInternalServerError
		return
	}
	fileName = filepath.Clean(fileName)
	dst := filepath.Join(d.dir, fileName)
	dst, err = filepath.Abs(dst)
	if err != nil {
		ctx.RespData = []byte("文件路径有误")
		ctx.RespStatusCode = http.StatusInternalServerError
		return
	}
	if !strings.Contains(dst, d.dir) {
		ctx.RespData = []byte("文件路径不合法")
		ctx.RespStatusCode = http.StatusInternalServerError
		return
	}
	_, err = os.Stat(dst)
	if err != nil {
		fmt.Println(err)
		ctx.RespData = []byte("文件不存在")
		ctx.RespStatusCode = http.StatusInternalServerError
		return
	}
	fn := filepath.Base(dst)
	header := ctx.Resp.Header()
	header.Set("Content-Disposition", "attachment;filename="+fn)
	header.Set("Content-Description", "File Transfer")
	header.Set("Content-Type", "application/octet-stream")
	header.Set("Content-Transfer-Encoding", "binary")
	header.Set("Expires", "0")
	header.Set("Cache-Control", "must-revalidate")
	header.Set("Pragma", "public")

	http.ServeFile(ctx.Resp, ctx.Req, dst)
}
