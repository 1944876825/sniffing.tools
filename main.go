package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"sniffing.tools/config"
	"sniffing.tools/server"
	"strconv"
	"strings"
	"time"
)

type UrlItemModel struct {
	Status  int // 2 进行中 1 完成 3 错误
	PlayUrl string
	TimeExp string
}

var Urls = make(map[string]UrlItemModel)

func main() {
	// 获取yaml配置
	config.Config.GetConfig()

	// 将gin设置为生产环境模式
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 嗅探
	r.GET("/xt", router)

	fmt.Println("API:", fmt.Sprintf("http://127.0.0.1:%d/xt?url=", config.Config.Port))
	err := r.Run(fmt.Sprintf(":%d", config.Config.Port))
	if err != nil {
		fmt.Println("启动失败", err)
		return
	}
}

// router 路由
func router(c *gin.Context) {
	url := c.Query("url")
	url = strings.TrimSpace(url)
	log.Println("url:", url)
	if len(url) == 0 {
		c.JSON(200, gin.H{"code": 404, "msg": "缺少URL"})
		return
	}
	if Urls[getMD5(url)].Status == 1 {
		timestamp := time.Now().Unix()
		timeExp, _ := strconv.ParseInt(Urls[getMD5(url)].TimeExp, 10, 64)
		if timestamp < timeExp {
			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "解析成功",
				"url":  Urls[getMD5(url)].PlayUrl,
			})
			return
		}
	}
	if Urls[getMD5(url)].Status == 2 {
		for Urls[getMD5(url)].Status == 2 {
			time.Sleep(time.Millisecond * 200)
		}
	} else {
		Urls[getMD5(url)] = UrlItemModel{
			Status:  2,
			PlayUrl: "",
			TimeExp: "",
		}
		go toParse(url)
		for Urls[getMD5(url)].Status == 2 {
			time.Sleep(time.Millisecond * 200)
		}
	}
	if Urls[getMD5(url)].PlayUrl != "" {
		c.JSON(200, gin.H{
			"code": 200,
			"msg":  "解析成功",
			"url":  Urls[getMD5(url)].PlayUrl,
		})
	} else {
		c.JSON(200, gin.H{
			"code": 404,
			"msg":  "解析失败",
		})
	}
}
func toParse(url string) {
	var mat = false
	var cuParse = config.ParseItemModel{}
	for _, parse := range config.Config.Parse {
		for _, match := range parse.Match {
			if strings.Contains(url, match) {
				mat = true
				break
			}
		}
		if mat {
			cuParse = parse
			break
		}
	}

	var ser = server.Model{}
	ser.Url = url
	ser.Data = cuParse

	if mat {
		if len(cuParse.Start) > 0 {
			ser.Url = cuParse.Start + ser.Url
		}
		if len(cuParse.End) > 0 {
			ser.Url = ser.Url + cuParse.End
		}
	} else {
		ser.Data.ContentType = []string{
			"application/vnd.apple.mpegurl",
			"video/mp4",
		}
	}
	ser.Init()
	playUrl, err := ser.StartFindResource()
	timestamp := time.Now().Unix()
	futureTimestamp := timestamp + config.Config.HcTime
	if err == nil {
		Urls[getMD5(url)] = UrlItemModel{
			Status:  1,
			PlayUrl: playUrl,
			TimeExp: strconv.FormatInt(futureTimestamp, 10),
		}
		return
	}
	log.Println("解析失败", err.Error())
	Urls[getMD5(url)] = UrlItemModel{
		Status:  3,
		PlayUrl: playUrl,
		TimeExp: strconv.FormatInt(futureTimestamp, 10),
	}
}

// getMD5 获取md5值
func getMD5(text string) string {
	// 创建一个MD5哈希对象
	hash := md5.New()
	// 将字符串转换为字节数组并计算哈希值
	hash.Write([]byte(text))
	hashValue := hash.Sum(nil)
	// 将哈希值转换为字符串表示
	return hex.EncodeToString(hashValue)
}
