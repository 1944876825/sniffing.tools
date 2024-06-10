package sniffing

// 缓存资源链接列表
// 格式 { md5(url): CacheUrlItemType }
var cacheUrls = make(map[string]*CacheUrlItemType)
