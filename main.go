package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"os/signal"
	"sniffing.tools/config"
	"sniffing.tools/server"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type UrlItemModel struct {
	Status  int // 2 进行中 1 完成 3 错误
	PlayUrl string
	TimeExp string
}

var Urls = make(map[string]UrlItemModel)

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 主函数
	go func() {
		<-signalChan
		server.CloseServers() // 确保关闭浏览器
		os.Exit(0)            // 退出程序
	}()
	fmt.Println("作者：By易仝 QQ：1944876825")
	fmt.Println("开源地址：https://github.com/1944876825/sniffing.tools")

	// 获取yaml配置
	config.Config.GetConfig()

	// 将gin设置为生产环境模式
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 嗅探
	r.GET("/xt", router)

	fmt.Println("程序启动成功 API:", fmt.Sprintf("http://127.0.0.1:%d/xt?url=", config.Config.Port))
	err := r.Run(fmt.Sprintf(":%d", config.Config.Port))
	if err != nil {
		fmt.Println("启动失败", err)
		return
	}
}

// router 路由
func router(c *gin.Context) {
	start := time.Now()
	url := c.Query("url")
	proxy := c.Query("proxy")
	url = strings.TrimSpace(url)
	md5Url := getMD5(url)
	log.Println("url: ", url)
	if len(url) < 1 {
		c.IndentedJSON(200, gin.H{"code": 404, "msg": "缺少URL"})
		return
	}
	if config.Config.Hc && Urls[md5Url].Status == 1 {
		timestamp := time.Now().Unix()
		timeExp, _ := strconv.ParseInt(Urls[md5Url].TimeExp, 10, 64)
		if timestamp < timeExp {
			end := time.Now()
			duration := end.Sub(start)
			c.IndentedJSON(200, gin.H{
				"code": 200,
				"msg":  "解析成功",
				"type": "缓存",
				"time": duration.Seconds(),
				"url":  Urls[md5Url].PlayUrl,
			})
			return
		}
	}
	if Urls[md5Url].Status == 2 {
		for Urls[md5Url].Status == 2 {
			time.Sleep(time.Millisecond * 200)
		}
	} else {
		Urls[md5Url] = UrlItemModel{
			Status:  2,
			PlayUrl: "",
			TimeExp: "",
		}
		go toParse(url, proxy)
		for Urls[md5Url].Status == 2 {
			time.Sleep(time.Millisecond * 200)
		}
	}
	end := time.Now()
	duration := end.Sub(start)
	if Urls[md5Url].PlayUrl != "" {
		c.IndentedJSON(200, gin.H{
			"code": 200,
			"msg":  "解析成功",
			"type": "最新",
			"time": duration.Seconds(),
			"url":  Urls[md5Url].PlayUrl,
		})
	} else {
		c.IndentedJSON(200, gin.H{
			"code": 404,
			"msg":  "解析失败",
			"time": duration.Seconds(),
		})
	}
}
func toParse(url, proxy string) {
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

	ser := server.GetServer()
	ser.Init(proxy)
	ser.Data = cuParse
	playUrl, err := ser.StartFindResource(url)

	timestamp := time.Now().Unix()
	futureTimestamp := timestamp + config.Config.HcTime
	md5Url := getMD5(url)
	if err == nil {
		Urls[md5Url] = UrlItemModel{
			Status:  1,
			PlayUrl: playUrl,
			TimeExp: strconv.FormatInt(futureTimestamp, 10),
		}
		return
	}
	log.Println("解析失败", err.Error())
	Urls[md5Url] = UrlItemModel{
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
