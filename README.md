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
  * match 资源特征匹配（可填多个）
  * wait 等待资源加载成功（可填多个，需要填写css选择器，如 .play #play，程序会等待这个元素加载完毕再开始嗅探）
  * click 如果有需要点击播放按钮才加载媒体的网页，可以填这个，不过需要手动使用css选择器找到对应按钮的（可填多个，当wait执行完才会执行这个）
  * contentType 要查找资源的 content-type 如：video/mp4
  * white 资源特征白名单（可填多个，包含此特征的资源会被选择）
  * black 资源特征白名单（可填多个，包含此特征的资源不会被选择）

## 使用方式
你的网址:端口/xt?url=

## LICENSE

[MIT](https://opensource.org/license/mit/)
