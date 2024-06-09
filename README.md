# sniffing.tools
嗅探网站网络资源工具

## 主要功能
* 嗅探网页中的媒体资源

## windows使用方法
- 直接双击启动
- 使用命令行启动
- 自制.bat脚本，文件名：start.bat，代码如下：
  ```
  start cmd /K "脚本名称.exe"
  ```

## linux使用方法
首先保证你的系统安装了谷歌浏览器，如果没有安装，请自行安装
- 运行方式1（缺点，命令行关闭就没了，可用于测试）
```
./脚本名称
```
- 运行方式2
```
nohup ./脚本名称
```
- 运行方式3
使用你习惯用的项目管理器，比如宝塔自带的Go项目管理器

## 使用方式
- 你的网址:端口/xt?url=
- 你的位置:端口/xt?proxy=你的代理&url= （proxy务必放在url前面）

## config.yaml 配置文件（不是很详细，请参考实例文件）
如果需要自定义配置，请将代码中的文件拷贝在程序同级目录，请参考config.yaml内注释

## LICENSE

[MIT](https://opensource.org/license/mit/)
