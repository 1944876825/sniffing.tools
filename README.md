# sniffing.tools
嗅探网页媒体资源

## 主要功能
* 嗅探网页中的媒体资源

## config.yaml 配置（不是很详细，请参考实例文件）

* port 项目运行端口
* headless 是否显示浏览器(false 当解析时，会显示浏览器加载网页的过程)
* hc_time 资源缓存时效（单位：秒，在此范围内，相同链接会走缓存的解析数据）
* parse
  * name 资源名称
  * match 资源特征匹配（多个）
  * start url前缀
  * end url后缀
  * suf 批量特征（多个，如.mp4用于嗅探获取媒体文件真实地址）
  * wait 等待（多个，需要填写css选择器，如 .play #play，程序会等待这个元素加载完毕再开始嗅探）
  * click 点击（多个，css选择器，如 需要点击播放按钮才加载媒体，就填这个）
  * black 屏蔽链接特征

## 使用方式
你的网址:端口/xt?url=

## LICENSE

[MIT](https://opensource.org/license/mit/)
