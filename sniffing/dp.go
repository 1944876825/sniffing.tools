package sniffing

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"path/filepath"
	"sniffing.tools/config"
	"strings"
	"sync"
	"time"
)

var mutex sync.Mutex

func New(p *config.ParseItemModel) *ChromeDp {
	// 如果浏览器数量大于最大数量，等待
	if len(Servers) >= config.Config.XcMax {
		for len(Servers) >= config.Config.XcMax {
			time.Sleep(time.Millisecond * 200)
		}
	}
	xc := &ChromeDp{
		data: p,
	}
	mutex.Lock()
	Servers = append(Servers, xc)
	mutex.Unlock()
	return xc
}

var Servers []*ChromeDp

type ChromeDp struct {
	data       *config.ParseItemModel
	ctx        context.Context
	playUrl    string
	needListen bool
	cancel     []context.CancelFunc
}

func (s *ChromeDp) Init(proxy string) {
	currentDir, _ := os.Getwd()
	dir := filepath.Join(currentDir, ".cache")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", config.Config.Headless), // 设置为true将在后台运行Chrome
		chromedp.Flag("user-data-dir", dir),               // 将缓存放在 .cache 目录下
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
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	s.cancel = append(s.cancel, cancel)
	// 创建新的Chrome浏览器上下文
	s.ctx, cancel = chromedp.NewContext(allocCtx)
	s.cancel = append(s.cancel, cancel)
	// 监听请求日志
	s.listen()

	// 设置窗口大小
	//_ = chromedp.Run(s.ctx, chromedp.EmulateViewport(1920, 1080))
	// 打开网页
	//_ = chromedp.Run(s.ctx, chromedp.Navigate("about:blank"))
}
func (s *ChromeDp) Run(url string) (string, error) {
	defer s.Cancel()
	s.needListen = true

	if len(s.data.White) < 1 {
		s.data.White = []string{".mp4", ".m3u8", ".flv"}
	}
	var task chromedp.Tasks
	if s.data.Headers != nil {
		log.Println("已设置header", s.data.Headers)
		//networkConditions := network.NewNetworkConditions().
		//	SetRequestHeaders(network.RequestHeaders{
		//		"referer": "http://example.com",
		//	})
		task = chromedp.Tasks{
			network.Enable(),
			network.SetExtraHTTPHeaders(network.Headers(map[string]interface{}{
				"X-Header": "my request header",
				"referer":  "https://jx.m3u8zy.top/",
			})),
		}
		//browserContext := chromedp.WithNewBrowserContext()
	}
	// 打开网页
	err := chromedp.Run(s.ctx,
		task,
		chromedp.Navigate(url))
	if err != nil {
		return "", err
	}

	// 等待网页加载完成
	if len(s.data.Wait) > 0 {
		go func() {
			if len(s.data.Click) > 0 {
				for _, click := range s.data.Click {
					_ = chromedp.Run(s.ctx, chromedp.Click(click))
				}
			}
		}()
	}

	var dpDone = false
	go func() {
		select {
		case <-s.ctx.Done():
			dpDone = true
		}
	}()
	var i = 0
	for {
		i++
		if i > config.Config.XtTime*20 {
			s.needListen = false
			log.Println("监听超时")
			break
		}
		if dpDone {
			log.Println("浏览器关闭")
			break
		}
		time.Sleep(time.Millisecond * 500)
	}
	if len(s.playUrl) != 0 {
		return s.playUrl, nil
	}
	return "", fmt.Errorf("解析失败")
}
func (s *ChromeDp) listen() {
	IsLogUrl := config.Config.IsLogUrl
	chromedp.ListenTarget(s.ctx, func(ev interface{}) {
		if s.needListen {
			switch ev := ev.(type) {
			case *network.EventRequestWillBeSent: // 发送
				req := ev.Request
				if IsLogUrl {
					log.Println("req url", req.URL)
				}
				for _, suf := range s.data.White {
					if strings.Contains(req.URL, suf) {
						if len(s.data.Black) > 0 {
							var isBlack = false
							for _, black := range s.data.Black {
								if len(black) != 0 && strings.Contains(req.URL, black) == true {
									isBlack = true
									break
								}
							}
							if isBlack == false {
								s.playUrl = req.URL
								s.needListen = false
							}
						} else {
							s.playUrl = req.URL
							s.needListen = false
						}
					}
				}
			case *network.EventResponseReceived: // 接收
				resp := ev.Response
				if IsLogUrl {
					log.Println("resp url", resp.URL)
				}
				for _, suf := range s.data.White {
					if strings.Contains(resp.URL, suf) {
						if len(s.data.Black) > 0 {
							var isBlack = false
							for _, black := range s.data.Black {
								if len(black) != 0 && strings.Contains(resp.URL, black) == true {
									isBlack = true
									break
								}
							}
							if isBlack == false {
								s.playUrl = resp.URL
								s.needListen = false
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
func (s *ChromeDp) Cancel() {
	removeServer(s)
	for _, cancelFunc := range s.cancel {
		cancelFunc()
	}
}

// 清除已经解析成功的ChromeDp实例
func removeServer(serverToRemove *ChromeDp) {
	for i, server := range Servers {
		if server == serverToRemove {
			Servers = append(Servers[:i], Servers[i+1:]...)
			break
		}
	}
}

// CloseServers 关闭所有ChromeDp实例
func CloseServers() {
	for _, xc := range Servers {
		xc.Cancel()
	}
	Servers = nil
}
