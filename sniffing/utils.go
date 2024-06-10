package sniffing

import (
	"sniffing.tools/utils"
	"sync"
)

var lock = sync.Mutex{}

func setCache(url string, value *CacheUrlItemType) {
	md5Url := utils.GetMD5(url)
	lock.Lock()
	if cacheUrls[md5Url] == nil {
		cacheUrls[md5Url] = value
	} else {
		cache := cacheUrls[md5Url]
		cache.Status = value.Status
		cache.ExpTime = value.ExpTime
		cache.PlayUrl = value.PlayUrl
	}
	lock.Unlock()
}
func getCache(url string) *CacheUrlItemType {
	md5Url := utils.GetMD5(url)
	return cacheUrls[md5Url]
}
