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
	"sniffing.tools/utils"
	"strings"
	"sync"
	"time"
)

var mutex sync.Mutex

func New() (*ChromeDp, error) {
	// 如果浏览器数量大于最大数量，等待
	if len(Servers) >= config.Config.XcMax {
		if config.Config.XcOut == 1 {
			return nil, fmt.Errorf("超出最大浏览器数量")
		}
		for len(Servers) >= config.Config.XcMax {
			time.Sleep(time.Millisecond * 1000)
		}
	}
	xc := &ChromeDp{
		finish: make(chan struct{}),
	}
	mutex.Lock()
	Servers = append(Servers, xc)
	mutex.Unlock()
	return xc, nil
}

var Servers []*ChromeDp

type ChromeDp struct {
	data       *config.ParseItemModel
	ctx        context.Context
	playUrl    string
	finish     chan struct{}
	needListen bool
	cancel     []context.CancelFunc
}

func (s *ChromeDp) SetData(data *config.ParseItemModel) {
	s.data = data
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
	if proxy == "" && config.Config.ProxyApi != "" {
		proxy = utils.GetProxy()
	}
	if proxy != "" {
		log.Println("使用代理：", proxy)
		opts = append(opts, chromedp.ProxyServer(proxy))
	}
	// 创建Chrome浏览器实例
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	s.cancel = append(s.cancel, cancel)
	// 创建新的Chrome浏览器上下文
	s.ctx, cancel = chromedp.NewContext(allocCtx)
	s.cancel = append(s.cancel, cancel)
	// 监听请求日志
	go s.listen()

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
	go func() {
		// 打开网页
		err := chromedp.Run(s.ctx,
			task,
			chromedp.Navigate(url))
		if err != nil {
			log.Println("网页加载失败", err)
			return
		}
		// 等待网页加载完成
		if len(s.data.Wait) > 0 {
			for _, wait := range s.data.Wait {
				_ = chromedp.Run(s.ctx, chromedp.WaitVisible(wait))
			}
		}
		// 点击页面元素
		if len(s.data.Click) > 0 {
			for _, click := range s.data.Click {
				_ = chromedp.Run(s.ctx, chromedp.Click(click))
			}
		}
	}()

	var dpDone = false
	go func() {
		select {
		case <-s.ctx.Done():
		case <-s.finish:
		}
		dpDone = true
	}()
	var i = 0
	for {
		if dpDone {
			break
		}
		i++
		if i > config.Config.XtTime {
			s.done("")
			log.Println("嗅探超时")
			break
		}
		time.Sleep(time.Second * 1)
	}
	if s.playUrl != "" {
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
					log.Println("send", req.URL)
				}
				for _, suf := range s.data.White {
					if strings.Contains(req.URL, suf) {
						var isBlack = false
						if len(s.data.Black) > 0 {
							for _, black := range s.data.Black {
								if len(black) != 0 && strings.Contains(req.URL, black) == true {
									isBlack = true
									break
								}
							}
						}
						if isBlack == false {
							s.done(req.URL)
							break
						}
					}
				}
			case *network.EventResponseReceived: // 接收
				resp := ev.Response
				if IsLogUrl {
					log.Println("recv", resp.URL)
				}
				for _, suf := range s.data.White {
					if strings.Contains(resp.URL, suf) {
						var isBlack = false
						if len(s.data.Black) > 0 {
							for _, black := range s.data.Black {
								if len(black) != 0 && strings.Contains(resp.URL, black) == true {
									isBlack = true
									break
								}
							}
						}
						if isBlack == false {
							s.done(resp.URL)
							break
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
func (s *ChromeDp) done(url string) {
	fmt.Println("done", url)
	s.playUrl = url
	s.needListen = false
	s.finish <- struct{}{}
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
