package storager

import (
	"context"
	"errors"
	"fmt"
	"go-liteflow/internal/pkg/log"
	"go-liteflow/internal/pkg/md5"
	"io"
	"os"
)

// 本地文件存储, 用于存储可执行文件
type Storager interface {
	// 写入可执行文件
	Write(ctx context.Context, ef []byte, hash string) (err error)
	// 读取可执行文件
	Read(ctx context.Context, hash string) (ef []byte, err error)
	// 获取可执行文件路径
	GetExecFilePath(hash string) string
}

type localStorager struct {
	path string
}

func NewStorager(ctx context.Context, path string) (Storager) {
	return &localStorager{
		path: path,
	}
}

// 将可执行文件写入到本地文件
func (ls *localStorager) Write(ctx context.Context, ef []byte, hash string) (err error) {
	if md5.Calc(ef) != hash {
		return errors.New("md5 not match")
	}

	err = os.MkdirAll(ls.path, 0755)
	if err != nil {
		return err
	}
	fpath := fmt.Sprintf("%s/%s", ls.path, hash)
	os.Chmod(fpath, 0755)
	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.Errorf("Open file failed. err: %v", err)
		return err
	}
	defer f.Close()

	_, err = f.Write(ef)
	if err != nil {
		log.Errorf("Write file failed. err: %v", err)
		return err
	}

	return nil
}

// 从本地文件读取可执行文件
func (ls *localStorager) Read(ctx context.Context, hash string) (ef []byte, err error) {

	f, err := os.Open(fmt.Sprintf("%s/%s", ls.path, hash))
	if err != nil {
		log.Errorf("Open file failed. err: %v", err)
		return nil, err
	}
	defer f.Close()

	ef, err = io.ReadAll(f)
	if err != nil {
		log.Errorf("Read file failed. err: %v", err)
		return nil, err
	}

	return ef, nil
}

func (ls *localStorager) GetExecFilePath(hash string) string {
	filepath := fmt.Sprintf("%s/%s", ls.path, hash)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return ""
	}
	return filepath
}
