package sniffing

// CacheUrlItemType 缓存资源链接类型
type CacheUrlItemType struct {
	Status  int    // 1 成功 2 进行中 3 失败
	PlayUrl string // 资源链接
	ExpTime string // 过期时间
}

// Status
const (
	Success = 1
	Loading = 2
	Failed  = 3
)

// ParseType type 缓存 最新
type ParseType struct {
	PlayUrl string
	Type    string
}
