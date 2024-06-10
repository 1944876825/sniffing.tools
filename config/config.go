package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

var Config = configModel{}

type configModel struct {
	Port       int    `yaml:"port"`
	Hc         bool   `yaml:"hc"`
	HcTime     int64  `yaml:"hc_time"`
	XtTime     int    `yaml:"xt_time"`
	XcOut      int    `yaml:"xc_out"`
	Headless   bool   `yaml:"headless"`
	Proxy      string `yaml:"proxy"`
	ProxyApi   string `yaml:"proxy_api"`
	XcMax      int    `yaml:"xc_max"`
	IsLogLocal bool   `yaml:"is_log_local"`
	IsLogUrl   bool   `yaml:"is_log_url"`
	Parse      []ParseItemModel
}
type ParseItemModel struct {
	Name        string                 `yaml:"name"`
	Match       []string               `yaml:"match"`
	Start       string                 `yaml:"start"`
	End         string                 `yaml:"end"`
	Wait        []string               `yaml:"wait"`
	Click       []string               `yaml:"click"`
	ContentType []string               `yaml:"contentType"`
	White       []string               `yaml:"white"`
	Black       []string               `yaml:"black"`
	Headers     map[string]interface{} `yaml:"headers"`
}

func (c *configModel) GetConfig() {
	content, err := os.ReadFile("./config.yaml")
	if err != nil {
		log.Println("配置文件读取失败", err)
		return
	}
	if err := yaml.Unmarshal(content, &c); err != nil {
		log.Println("配置文件解析失败", err)
		return
	}
}
