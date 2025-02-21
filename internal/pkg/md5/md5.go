package md5

import (
	"crypto/md5"
	"fmt"
)

func Calc(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}
