package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"sniffing.tools/config"
	"sniffing.tools/server"
	"strconv"
	"strings"
	"time"
)

type UrlItemModel struct {
	Status  string
	PlayUrl string
	TimeExp string
}

var Urls = make(map[string]UrlItemModel)

func main() {
	config.Config.GetConfig()
	r := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	r.GET("/xt", router)
	err := r.Run(fmt.Sprintf(":%d", config.Config.Port))
	if err != nil {
		return
	}
}
func router(c *gin.Context) {
	url := c.Query("url")
	url = strings.TrimSpace(url)
	if len(url) == 0 {
		c.JSON(200, gin.H{"code": 404, "msg": "解析失败"})
		return
	}
	fmt.Println(getMD5(url))
	if Urls[getMD5(url)].Status == "success" {
		timestamp := time.Now().Unix()
		timeExp, _ := strconv.ParseInt(Urls[getMD5(url)].TimeExp, 10, 64)
		fmt.Println(timestamp, timeExp, timestamp > timeExp)
		if timestamp < timeExp {
			c.JSON(200, gin.H{
				"code": 200,
				"msg":  "解析成功",
				"url":  Urls[getMD5(url)].PlayUrl,
			})
			return
		}
	}
	Urls[getMD5(url)] = UrlItemModel{
		Status:  "start",
		PlayUrl: "",
		TimeExp: "",
	}
	go toParse(url, c)
	for Urls[getMD5(url)].Status == "start" {
		time.Sleep(time.Millisecond * 100)
	}
}
func toParse(url string, c *gin.Context) {
	var mat = false
	for _, parse := range config.Config.Parse {
		for _, match := range parse.Match {
			if strings.Contains(url, match) {
				mat = true
				break
			}
		}
		if mat {
			var ser = server.ServerModel{}
			ser.Data = parse
			ser.Url = url
			if len(parse.Start) != 0 {
				ser.Url = parse.Start + ser.Url
			}
			if len(parse.End) != 0 {
				ser.Url = ser.Url + parse.End
			}
			code, msg, playUrl := ser.Xt()
			if code == 200 {
				timestamp := time.Now().Unix()
				futureTimestamp := timestamp + config.Config.HcTime
				Urls[getMD5(url)] = UrlItemModel{
					Status:  "success",
					PlayUrl: playUrl,
					TimeExp: strconv.FormatInt(futureTimestamp, 10),
				}
				c.JSON(200, gin.H{
					"code": code,
					"msg":  msg,
					"url":  playUrl,
				})
				break
			}
		}
	}
	Urls[getMD5(url)] = UrlItemModel{
		Status: "error",
	}
	c.JSON(200, gin.H{
		"code": 404,
		"msg":  "解析失败",
		"url":  "",
	})
}
func getMD5(text string) string {
	// 创建一个MD5哈希对象
	hash := md5.New()
	// 将字符串转换为字节数组并计算哈希值
	hash.Write([]byte(text))
	hashValue := hash.Sum(nil)
	// 将哈希值转换为字符串表示
	return hex.EncodeToString(hashValue)
}
