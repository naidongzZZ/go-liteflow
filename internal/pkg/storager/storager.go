package storager

import (
	"context"
	"errors"
	"fmt"
	"go-liteflow/internal/pkg/md5"
	"io"
	"log/slog"
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

	f, err := os.OpenFile(fmt.Sprintf("%s/%s", ls.path, hash), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("Open file failed.", slog.Any("err", err))
		return err
	}
	defer f.Close()

	_, err = f.Write(ef)
	if err != nil {
		slog.Error("Write file failed.", slog.Any("err", err))
		return err
	}

	return nil
}

// 从本地文件读取可执行文件
func (ls *localStorager) Read(ctx context.Context, hash string) (ef []byte, err error) {

	f, err := os.Open(fmt.Sprintf("%s/%s", ls.path, hash))
	if err != nil {
		slog.Error("Open file failed.", slog.Any("err", err))
		return nil, err
	}
	defer f.Close()

	ef, err = io.ReadAll(f)
	if err != nil {
		slog.Error("Read file failed.", slog.Any("err", err))
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
