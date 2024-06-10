package sniffing

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"sniffing.tools/config"
	"sniffing.tools/utils"
	"strconv"
	"strings"
	"time"
)

func Xt(c *gin.Context) {
	start := time.Now()
	//url := c.Query("url")
	url := utils.DealUrl(c.Request.URL.String())
	proxy := c.Query("proxy")
	log.Println(fmt.Sprintf("url: %s proxy: %s", url, proxy))

	url = strings.TrimSpace(url)
	proxy = strings.TrimSpace(proxy)

	if len(url) < 1 {
		c.IndentedJSON(200, gin.H{"code": 404, "msg": "缺少URL"})
		return
	}

	cache := getCache(url)
	if cache == nil { // 新
		cache = &CacheUrlItemType{
			Status:  Loading,
			PlayUrl: "",
			ExpTime: "",
		}
		setCache(url, cache)
		go toParse(url, proxy)
	}
	if config.Config.Hc {
		if cache.Status == Success { // 已有缓存
			timestamp := time.Now().Unix()
			timeExp, _ := strconv.ParseInt(cache.ExpTime, 10, 64)
			if timestamp < timeExp {
				end := time.Now()
				duration := end.Sub(start)
				c.IndentedJSON(200, gin.H{
					"code": 200,
					"msg":  "解析成功",
					"type": "缓存",
					"time": duration.Seconds(),
					"url":  cache.PlayUrl,
				})
				return
			}
		}
	}
	if cache.Status == Failed {
		cache.Status = Loading
		go toParse(url, proxy)
	}
	for cache.Status == Loading { // 等待解析成功
		time.Sleep(time.Millisecond * 500)
	}
	end := time.Now()
	duration := end.Sub(start) // 计算解析耗时

	if cache.PlayUrl != "" {
		c.IndentedJSON(200, gin.H{
			"code": 200,
			"msg":  "解析成功",
			"type": "最新",
			"time": duration.Seconds(),
			"url":  cache.PlayUrl,
		})
	} else {
		c.IndentedJSON(200, gin.H{
			"code": 404,
			"msg":  "解析失败",
			"time": duration.Seconds(),
		})
	}
}
func toParse(url string, proxy string) {
	// 匹配资源解析配置
	var isMatchSucc = false
	var cuParse = config.ParseItemModel{}
	for _, parse := range config.Config.Parse {
		for _, match := range parse.Match {
			if strings.Contains(url, match) {
				isMatchSucc = true
				break
			}
		}
		if isMatchSucc {
			cuParse = parse
			break
		}
	}

	ser := New(&cuParse)
	ser.Init(proxy)
	playUrl, err := ser.Run(url)

	timestamp := time.Now().Unix()
	futureTimestamp := timestamp + config.Config.HcTime

	if err != nil {
		log.Println("解析失败", err.Error())
		setCache(url, &CacheUrlItemType{
			Status:  Failed,
			PlayUrl: playUrl,
			ExpTime: strconv.FormatInt(futureTimestamp, 10),
		})
		return
	}
	setCache(url, &CacheUrlItemType{
		Status:  Success,
		PlayUrl: playUrl,
		ExpTime: strconv.FormatInt(futureTimestamp, 10),
	})
}
