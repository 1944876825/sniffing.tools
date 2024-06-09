package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"os/signal"
	"sniffing.tools/config"
	"sniffing.tools/sniffing"
	"sniffing.tools/utils"
	"syscall"
)

func main() {
	fmt.Println("作者：By易仝 QQ：1944876825")
	fmt.Println("开源地址：https://github.com/1944876825/sniffing.tools")
	gin.SetMode(gin.ReleaseMode)
	config.Config.GetConfig()

	if config.Config.IsLogLocal {
		utils.OpenLogLocal()
	}

	quit()
	startGin()
}

func startGin() {
	r := gin.Default()
	r.GET("/xt", sniffing.Xt) // 嗅探
	fmt.Println("程序启动成功 API:", fmt.Sprintf("http://127.0.0.1:%d/xt?url=", config.Config.Port))
	err := r.Run(fmt.Sprintf(":%d", config.Config.Port))
	if err != nil {
		fmt.Println("启动失败", err)
		return
	}
}

func quit() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		sniffing.CloseServers()
		utils.CloseLogLocal()
		os.Exit(0) // 退出程序
	}()
}
