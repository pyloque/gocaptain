package gocaptain

type LocalKv struct {
	GlobalVersion int64
	Versions      map[string]int64
	Kvs           map[string]map[string]interface{}
}

func NewLocalKv() *LocalKv {
	return &LocalKv{-1, map[string]int64{}, map[string]map[string]interface{}{}}
}

func (this *LocalKv) GetVersion(key string) int64 {
	value, ok := this.Versions[key]
	if ok {
		return value
	}
	return -1
}

func (this *LocalKv) SetVersion(key string, version int64) {
	this.Versions[key] = version
}

func (this *LocalKv) GetKv(key string) map[string]interface{} {
	value, ok := this.Kvs[key]
	if ok {
		return value
	}
	return map[string]interface{}{}
}

func (this *LocalKv) ReplaceKv(key string, value map[string]interface{}) {
	this.Kvs[key] = value
}

func (this *LocalKv) InitKv(key string) {
	this.Kvs[key] = map[string]interface{}{}
}
