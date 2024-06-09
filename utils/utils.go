package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// GetMD5 获取字符串md5值
func GetMD5(text string) string {
	// 创建一个MD5哈希对象
	hash := md5.New()
	// 将字符串转换为字节数组并计算哈希值
	hash.Write([]byte(text))
	hashValue := hash.Sum(nil)
	// 将哈希值转换为字符串表示
	return hex.EncodeToString(hashValue)
}

// DealUrl 获取url参数
func DealUrl(url string) string {
	split := strings.SplitN(url, "url=", 2)
	return split[1]
}
