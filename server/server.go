package server

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"sniffing.tools/config"
	"strings"
	"sync"
	"time"
)

var Wg sync.WaitGroup

type ServerModel struct {
	Data       config.ParseItemModel
	Url        string
	playUrl    string
	ctx        context.Context
	needListen bool
	cancel     []context.CancelFunc
}

func (s *ServerModel) Init() {
	// 创建Chrome浏览器上下文
	var cancel context.CancelFunc
	s.ctx, cancel = chromedp.NewContext(context.Background())
	s.cancel = append(s.cancel, cancel)
	// 截图选项
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", config.Config.Headless), // 设置为true将在后台运行Chrome
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	// 创建Chrome浏览器实例
	var allocCtx context.Context
	allocCtx, cancel = chromedp.NewExecAllocator(s.ctx, opts...)
	s.cancel = append(s.cancel, cancel)
	// 创建新的Chrome浏览器上下文
	s.ctx, cancel = chromedp.NewContext(allocCtx)
	s.cancel = append(s.cancel, cancel)
	// 设置窗口大小
	err := chromedp.Run(s.ctx, chromedp.EmulateViewport(1920, 1080))
	if err != nil {
		fmt.Println("设置窗口大小时出错:", err)
		return
	}
}
func (s *ServerModel) Xt() (int, string, string) {
	s.Init()
	s.playUrl = ""
	// 导航到网页
	fmt.Println("网页:", s.Url)
	err := chromedp.Run(s.ctx, chromedp.Navigate(s.Url))
	if err != nil {
		fmt.Println("导航到网页时出错:", err)
		s.Cancel()
		return 404, "导航到网页时出错", ""
	}

	if len(s.Data.Wait) > 0 {
		//fmt.Println("等待网页加载完成")
		for _, wait := range s.Data.Wait {
			err = chromedp.Run(s.ctx, chromedp.WaitVisible(wait))
			if err != nil {
				fmt.Println("等待网页加载完成时出错:", err)
				s.Cancel()
				return 404, "等待网页加载完成时出错", ""
			}
		}
		//fmt.Println("网页加载完成")
	} else {
		err = chromedp.Run(s.ctx, chromedp.WaitVisible("body"))
		if err != nil {
			fmt.Println("等待网页加载完成时出错:", err)
			s.Cancel()
			return 404, "等待网页加载完成时出错", ""
		}
	}
	if len(s.Data.Click) > 0 {
		//fmt.Println("等待点击完成")
		for _, click := range s.Data.Click {
			err = chromedp.Run(s.ctx, chromedp.Click(click))
			if err != nil {
				fmt.Println("点击网页元素失败:", err)
				s.Cancel()
				return 404, "点击网页元素失败", ""
			}
		}
		//fmt.Println("点击完成")
	}

	//fmt.Println("监听请求日志")
	s.needListen = true
	s.listenForNetworkEvent()

	var i = 0
	for s.playUrl == "" {
		i++
		if i > 300 {
			s.needListen = false
			fmt.Println("监听超时")
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	s.Cancel()
	if len(s.playUrl) != 0 {
		fmt.Println("play", s.playUrl)
		return 200, "解析成功", s.playUrl
	}
	return 404, "解析失败", ""
}
func (s *ServerModel) listenForNetworkEvent() {
	chromedp.ListenTarget(s.ctx, func(ev interface{}) {
		if s.needListen {
			switch ev := ev.(type) {
			case *network.EventResponseReceived:
				resp := ev.Response
				if len(resp.Headers) != 0 {
					for _, suf := range s.Data.Suf {
						if strings.Contains(resp.URL, suf) {
							if len(s.Data.Black) > 0 {
								for _, black := range s.Data.Black {
									if len(black) != 0 && strings.Contains(resp.URL, black) == false {
										s.playUrl = resp.URL
										s.needListen = false
									}
								}
							} else {
								s.playUrl = resp.URL
								s.needListen = false
							}
						}
					}
				}
			}
		}
	})
}

func (s *ServerModel) Cancel() {
	for _, cancelFunc := range s.cancel {
		cancelFunc()
	}
}
