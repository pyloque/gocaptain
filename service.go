package gocaptain

import (
	"fmt"
	"math/rand"
)

const ORIGIN_DEFAULT_PROBE = 1024

type ServiceItem struct {
	Host  string
	Port  int
	Ttl   int
	Probe int
}

type LocalService struct {
	GlobalVersion int64
	Versions      map[string]int64
	ServiceLists  map[string][]*ServiceItem
	failovers     map[string][]*ServiceItem
}

func NewServiceItem(host string, port int) *ServiceItem {
	return NewServiceItemWithTtl(host, port, 30)
}

func NewServiceItemWithTtl(host string, port int, ttl int) *ServiceItem {
	return &ServiceItem{host, port, ttl, ORIGIN_DEFAULT_PROBE}
}

func (this *ServiceItem) UrlRoot() string {
	return fmt.Sprintf("http://%v:%v", this.Host, this.Port)
}

func NewLocalService() *LocalService {
	return &LocalService{-1, map[string]int64{}, map[string][]*ServiceItem{}, map[string][]*ServiceItem{}}
}

func (this *LocalService) RandomService(name string) *ServiceItem {
	services := this.ServiceLists[name]
	failovers := this.failovers[name]
	if len(services) == 0 {
		if failovers == nil || len(failovers) == 0 {
			panic(&CaptainError{"no service provided"})
		}
		services = failovers
	}
	ind := rand.Intn(len(services))
	return services[ind]
}

func (this *LocalService) GetVersion(name string) int64 {
	version, ok := this.Versions[name]
	if !ok {
		version = -1
	}
	return version
}

func (this *LocalService) SetVersion(name string, version int64) {
	this.Versions[name] = version
}

func (this *LocalService) InitService(name string) {
	this.ServiceLists[name] = []*ServiceItem{}
}

func (this *LocalService) ReplaceService(name string, services []*ServiceItem) {
	this.ServiceLists[name] = services
}

func (this *LocalService) Failover(name string, services []*ServiceItem) {
	this.failovers[name] = services
}
