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
	md5Url := utils.GetMD5(url)

	if len(url) < 1 {
		c.IndentedJSON(200, gin.H{"code": 404, "msg": "缺少URL"})
		return
	}
	if config.Config.Hc && cacheUrlList[md5Url].Status == Success { // 已有缓存
		timestamp := time.Now().Unix()
		timeExp, _ := strconv.ParseInt(cacheUrlList[md5Url].ExpTime, 10, 64)
		if timestamp < timeExp {
			end := time.Now()
			duration := end.Sub(start)
			c.IndentedJSON(200, gin.H{
				"code": 200,
				"msg":  "解析成功",
				"type": "缓存",
				"time": duration.Seconds(),
				"url":  cacheUrlList[md5Url].PlayUrl,
			})
			return
		}
	}

	if cacheUrlList[md5Url].Status == Loading { // 正在解析
		for cacheUrlList[md5Url].Status == 2 { // 判断是否解析成功，否则一直循环等待
			time.Sleep(time.Millisecond * 200)
		}
	} else { // 未解析
		cacheUrlList[md5Url] = CacheUrlItemType{
			Status:  2,
			PlayUrl: "",
			ExpTime: "",
		}
		go toParse(url, proxy)
		for cacheUrlList[md5Url].Status == Loading {
			time.Sleep(time.Millisecond * 200)
		}
	}
	end := time.Now()
	duration := end.Sub(start) // 计算解析耗时

	if cacheUrlList[md5Url].PlayUrl != "" {
		c.IndentedJSON(200, gin.H{
			"code": 200,
			"msg":  "解析成功",
			"type": "最新",
			"time": duration.Seconds(),
			"url":  cacheUrlList[md5Url].PlayUrl,
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
	md5Url := utils.GetMD5(url)
	if err == nil {
		cacheUrlList[md5Url] = CacheUrlItemType{
			Status:  1,
			PlayUrl: playUrl,
			ExpTime: strconv.FormatInt(futureTimestamp, 10),
		}
		return
	}
	log.Println("解析失败", err.Error())
	cacheUrlList[md5Url] = CacheUrlItemType{
		Status:  3,
		PlayUrl: playUrl,
		ExpTime: strconv.FormatInt(futureTimestamp, 10),
	}
}
