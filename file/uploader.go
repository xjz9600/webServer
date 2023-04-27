package file

import (
	"github.com/google/uuid"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	web "routing"
)

type fileUploader struct {
	fileField   string
	dstPathFunc func(header *multipart.FileHeader) string
}

type uploadOptions func(*fileUploader)

func NewFileUploader(opts ...uploadOptions) *fileUploader {
	res := &fileUploader{
		fileField: "file",
		dstPathFunc: func(header *multipart.FileHeader) string {
			return filepath.Join("testdata", "upload", uuid.New().String())
		},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func UploaderWithFileField(fileField string) uploadOptions {
	return func(uploader *fileUploader) {
		uploader.fileField = fileField
	}
}

func UploaderWithDstPathFunc(dstPathFunc func(header *multipart.FileHeader) string) uploadOptions {
	return func(uploader *fileUploader) {
		uploader.dstPathFunc = dstPathFunc
	}
}

func (f *fileUploader) Handle(ctx *web.Context) {
	file, fileHeader, err := ctx.Req.FormFile(f.fileField)
	if err != nil {
		ctx.RespData = []byte("上传失败" + err.Error())
		ctx.RespStatusCode = http.StatusInternalServerError
		return
	}
	defer file.Close()
	dst := f.dstPathFunc(fileHeader)
	dir, _ := path.Split(dst)
	// 创建路径上的所有路径
	os.MkdirAll(dir, os.ModePerm)
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		ctx.RespData = []byte("创建文件失败" + err.Error())
		ctx.RespStatusCode = http.StatusInternalServerError
		return
	}
	defer dstFile.Close()
	_, err = io.CopyBuffer(dstFile, file, nil)
	if err != nil {
		ctx.RespData = []byte("写入文件失败" + err.Error())
		ctx.RespStatusCode = http.StatusInternalServerError
		return
	}
	ctx.RespStatusCode = http.StatusOK
	ctx.RespData = []byte("上传成功")
}
