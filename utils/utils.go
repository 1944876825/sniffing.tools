package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"sniffing.tools/config"
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

func GetProxy() string {
	res, err := Get(config.Config.ProxyApi)
	if err != nil {
		log.Println("代理获取失败", err.Error())
		return ""
	}
	return string(res)
}

func Get(url string) ([]byte, error) {
	req, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	res, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	return res, nil
}
