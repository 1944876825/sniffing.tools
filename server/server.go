package server

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"sniffing.tools/config"
	"strings"
	"time"
)

type Model struct {
	Data       config.ParseItemModel
	Url        string
	ctx        context.Context
	playUrl    string
	needListen bool
	cancel     []context.CancelFunc
}

func (s *Model) Init() {
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
	_ = chromedp.Run(s.ctx, chromedp.EmulateViewport(1920, 1080))
}
func (s *Model) StartFindResource() (string, error) {
	defer s.Cancel()
	if len(s.Data.White) < 1 {
		s.Data.White = []string{".mp4", ".m3u8", ".flv"}
	}
	// 监听请求日志
	s.needListen = true
	s.listenForNetworkEvent()

	// 打开网页
	err := chromedp.Run(s.ctx, chromedp.Navigate(s.Url))
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
		if i > config.Config.XtTime*5 {
			s.needListen = false
			fmt.Println("监听超时")
			break
		}
		time.Sleep(time.Millisecond * 200)
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
	for _, cancelFunc := range s.cancel {
		cancelFunc()
	}
}
