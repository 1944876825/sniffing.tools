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

	parse, err := toParse(url, proxy)
	end := time.Now()
	duration := end.Sub(start) // 计算解析耗时
	if err != nil {
		log.Println("res err", err.Error())
		c.IndentedJSON(200, gin.H{
			"code": 404,
			"msg":  "解析失败，" + err.Error(),
			"time": duration.Seconds(),
		})
		return
	}
	log.Println("res success", parse.PlayUrl)
	c.IndentedJSON(200, gin.H{
		"code": 200,
		"msg":  "解析成功",
		"type": parse.Type,
		"time": duration.Seconds(),
		"url":  parse.PlayUrl,
	})
}

func toParse(url string, proxy string) (*ParseType, error) {
	cache := getCache(url)
	if cache == nil { // 新
		cache = &CacheUrlItemType{
			Status:  Loading,
			PlayUrl: "",
			ExpTime: "",
		}
		setCache(url, cache)
	}
	if config.Config.Hc {
		if cache.Status == Success { // 已有缓存
			timestamp := time.Now().Unix()
			timeExp, _ := strconv.ParseInt(cache.ExpTime, 10, 64)
			if timestamp < timeExp {
				return &ParseType{
					PlayUrl: cache.PlayUrl,
					Type:    "缓存",
				}, nil
			}
		}
	}

	ser, err := New()
	if err != nil {
		return nil, err
	}
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
	ser.SetData(&cuParse)
	ser.Init(proxy)
	playUrl, err := ser.Run(url)

	timestamp := time.Now().Unix()
	futureTimestamp := timestamp + config.Config.HcTime

	if err != nil {
		setCache(url, &CacheUrlItemType{
			Status:  Failed,
			PlayUrl: "",
			ExpTime: strconv.FormatInt(futureTimestamp, 10),
		})
		return nil, err
	}
	setCache(url, &CacheUrlItemType{
		Status:  Success,
		PlayUrl: playUrl,
		ExpTime: strconv.FormatInt(futureTimestamp, 10),
	})
	return &ParseType{
		PlayUrl: playUrl,
		Type:    "最新",
	}, nil
}
