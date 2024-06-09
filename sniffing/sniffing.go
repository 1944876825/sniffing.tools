package sniffing

// 缓存资源链接列表
// 格式 { md5(url): CacheUrlItemType }
var cacheUrlList = make(map[string]CacheUrlItemType)
