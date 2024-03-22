package server

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"sniffing.tools/config"
	"strings"
	"sync"
	"time"
)

var mutex sync.Mutex

func GetServer() *Model {
	if len(Servers) >= config.Config.XcMax {
		for len(Servers) >= config.Config.XcMax {
			time.Sleep(time.Millisecond * 200)
		}
	}
	xc := &Model{}
	mutex.Lock()
	Servers = append(Servers, xc)
	mutex.Unlock()
	return xc
}

var Servers []*Model

type Model struct {
	Data       config.ParseItemModel
	ctx        context.Context
	playUrl    string
	needListen bool
	cancel     []context.CancelFunc
}

func (s *Model) Init(proxy string) {
	// 创建Chrome浏览器上下文
	//var cancel context.CancelFunc
	//s.ctx, cancel = chromedp.NewContext(context.Background())
	//s.cancel = append(s.cancel, cancel)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", config.Config.Headless), // 设置为true将在后台运行Chrome
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)
	if proxy == "" && config.Config.Proxy != "" {
		proxy = config.Config.Proxy
	}
	if proxy != "" {
		log.Println("使用代理：", proxy, "访问")
		opts = append(opts, chromedp.ProxyServer(proxy))
	}
	// 创建Chrome浏览器实例
	//var allocCtx context.Context
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	s.cancel = append(s.cancel, cancel)
	// 创建新的Chrome浏览器上下文
	s.ctx, cancel = chromedp.NewContext(allocCtx)
	s.cancel = append(s.cancel, cancel)
	// 监听请求日志
	s.listenForNetworkEvent()

	// 设置窗口大小
	//_ = chromedp.Run(s.ctx, chromedp.EmulateViewport(1920, 1080))
	// 打开网页
	//_ = chromedp.Run(s.ctx, chromedp.Navigate("about:blank"))
}
func (s *Model) StartFindResource(url string) (string, error) {
	defer s.Cancel()
	s.needListen = true

	if len(s.Data.White) < 1 {
		s.Data.White = []string{".mp4", ".m3u8", ".flv"}
	}

	// 打开网页
	err := chromedp.Run(s.ctx, chromedp.Navigate(url))
	if err != nil {
		return "", err
	}

	// 等待网页加载完成
	if len(s.Data.Wait) > 0 {
		go func() {
			if len(s.Data.Click) > 0 {
				for _, click := range s.Data.Click {
					_ = chromedp.Run(s.ctx, chromedp.Click(click))
				}
			}
		}()
	}

	var i = 0
	for s.playUrl == "" {
		i++
		if i > config.Config.XtTime*20 {
			s.needListen = false
			fmt.Println("监听超时")
			break
		}
		time.Sleep(time.Millisecond * 50)
	}
	if len(s.playUrl) != 0 {
		return s.playUrl, nil
	}
	return "", fmt.Errorf("解析失败")
}
func (s *Model) listenForNetworkEvent() {
	chromedp.ListenTarget(s.ctx, func(ev interface{}) {
		if s.needListen {
			switch ev := ev.(type) {
			case *network.EventRequestWillBeSent:
				req := ev.Request
				//log.Println("req url", req.URL)
				for _, suf := range s.Data.White {
					if strings.Contains(req.URL, suf) {
						if len(s.Data.Black) > 0 {
							for _, black := range s.Data.Black {
								if len(black) != 0 && strings.Contains(req.URL, black) == false {
									s.playUrl = req.URL
									s.needListen = false
								}
							}
						} else {
							s.playUrl = req.URL
							s.needListen = false
						}
					}
				}
			case *network.EventResponseReceived:
				resp := ev.Response
				//log.Println("resp url", resp.URL)
				for _, suf := range s.Data.White {
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
	})
}
func (s *Model) Cancel() {
	log.Println("len1", len(Servers))
	removeServer(s)
	log.Println("len2", len(Servers))
	for _, cancelFunc := range s.cancel {
		cancelFunc()
	}
}
func removeServer(serverToRemove *Model) {
	for i, server := range Servers {
		if server == serverToRemove {
			Servers = append(Servers[:i], Servers[i+1:]...)
			break
		}
	}
}
func CloseServers() {
	for _, xc := range Servers {
		xc.Cancel()
	}
	Servers = nil
}
