package gocaptain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type CaptainError struct {
	reason string
}

func (this *CaptainError) Error() string {
	return this.reason
}

func SilentOnPanic() {
	if err := recover(); err != nil {
		log.Printf("%v", err)
	}
}

type IServiceObserver interface {
	Online(client *CaptainClient, name string)
	AllOnline(client *CaptainClient)
	Offline(client *CaptainClient, name string)
}

type CaptainClient struct {
	origins   []*ServiceItem
	locals    *LocalService
	provided  map[string]*ServiceItem
	watched   map[string]bool
	failovers map[string][]*ServiceItem
	urlRoot   string
	observers []IServiceObserver
	keeper    *ServiceKeeper
	waiter    chan bool
}

func NewCaptainClient(host string, port int) *CaptainClient {
	return NewCaptainClientWithOrigins(NewServiceItem(host, port))
}

func NewCaptainClientWithOrigins(origins ...*ServiceItem) *CaptainClient {
	client := &CaptainClient{
		origins,
		NewLocalService(),
		map[string]*ServiceItem{},
		map[string]bool{},
		map[string][]*ServiceItem{},
		"",
		[]IServiceObserver{},
		nil,
		nil,
	}
	keeper := NewServiceKeeper(client)
	client.keeper = keeper
	return client
}

func (this *CaptainClient) ShuffleUrlRoot() {
	ind := rand.Intn(len(this.origins))
	this.urlRoot = this.origins[ind].UrlRoot()
}

func (this *CaptainClient) UrlRoot() string {
	if this.urlRoot == "" {
		this.ShuffleUrlRoot()
	}
	return this.urlRoot
}

func (this *CaptainClient) CheckDirty() bool {
	url := fmt.Sprintf(
		"%v/api/service/dirty?version=%v",
		this.urlRoot, this.locals.GlobalVersion)
	resp, err := http.Get(url)
	if err != nil {
		panic(&CaptainError{fmt.Sprintf("call api dirty error:%v", err.Error())})
	}
	decoder := json.NewDecoder(resp.Body)
	var data struct {
		Ok      bool
		Dirty   bool
		Version uint64
	}
	err = decoder.Decode(&data)
	if err != nil {
		panic(&CaptainError{fmt.Sprintf("api dirty response illegal:%v", err)})
	}
	return data.Dirty
}

func (this *CaptainClient) CheckVersions() []string {
	dirties := []string{}
	if len(this.watched) == 0 {
		return dirties
	}
	var buf bytes.Buffer
	var i = 0
	for name := range this.watched {
		buf.WriteString("name=" + name)
		if i < len(this.watched)-1 {
			buf.WriteString("&")
		}
	}
	url := fmt.Sprintf("%v/api/service/version?%v", this.urlRoot, buf.String())
	resp, err := http.Get(url)
	if err != nil {
		panic(&CaptainError{fmt.Sprintf("call api check version error:%v", err)})
	}
	decoder := json.NewDecoder(resp.Body)
	var data struct {
		Ok       bool
		Versions map[string]int64
	}
	err = decoder.Decode(&data)
	if err != nil {
		panic(&CaptainError{fmt.Sprintf("api check version response illegal:%s", err)})
	}
	for name, version := range data.Versions {
		if version != this.locals.GetVersion(name) {
			dirties = append(dirties, name)
		}
	}
	return dirties
}

func (this *CaptainClient) ReloadService(name string) {
	url := fmt.Sprintf("%v/api/service/set?name=%v", this.urlRoot, name)
	resp, err := http.Get(url)
	if err != nil {
		panic(&CaptainError{fmt.Sprintf("call api service set error:%v", err)})
	}
	decoder := json.NewDecoder(resp.Body)
	var data struct {
		Ok       bool
		Version  int64
		Services []*ServiceItem
	}
	err = decoder.Decode(&data)
	if err != nil {
		panic(&CaptainError{fmt.Sprintf("api service set response illegal:%v", err)})
	}
	this.locals.SetVersion(name, data.Version)
	this.locals.ReplaceService(name, data.Services)
	if len(data.Services) == 0 && this.IsHealthy(name) {
		this.Offline(name)
	} else if len(data.Services) != 0 && !this.IsHealthy(name) {
		this.Online(name)
	}
}

func (this *CaptainClient) KeepService() {
	for name, item := range this.provided {
		url := fmt.Sprintf(
			"%v/api/service/keep?name=%v&host=%v&port=%v&ttl=%v",
			this.urlRoot, name, item.Host, item.Port, item.Ttl)
		_, err := http.Get(url)
		if err != nil {
			panic(&CaptainError{fmt.Sprintf("call api keep service error:%v", err)})
		}
	}
}

func (this *CaptainClient) CancelService() {
	for name, item := range this.provided {
		url := fmt.Sprintf(
			"%v/api/service/cancel?name=%v&host=%v&port=%v",
			this.urlRoot, name, item.Host, item.Port)
		_, err := http.Get(url)
		if err != nil {
			panic(&CaptainError{fmt.Sprintf("call api cancel service error:%v", err)})
		}
	}
}

func (this *CaptainClient) Watch(names ...string) *CaptainClient {
	for _, name := range names {
		this.watched[name] = false
		this.locals.InitService(name)
	}
	return this
}

func (this *CaptainClient) Failover(name string, items ...*ServiceItem) *CaptainClient {
	this.failovers[name] = items
	return this
}

func (this *CaptainClient) Provide(name string, service *ServiceItem) *CaptainClient {
	this.provided[name] = service
	return this
}

func (this *CaptainClient) Select(name string) *ServiceItem {
	return this.locals.RandomService(name, this.failovers[name])
}

func (this *CaptainClient) KeepAlive(keepAlive int64) *CaptainClient {
	this.keeper.KeepAlive = keepAlive
	return this
}

func (this *CaptainClient) CheckInterval(interval int64) *CaptainClient {
	this.keeper.CheckInterval = time.Duration(interval)
	return this
}

func (this *CaptainClient) Observe(observer IServiceObserver) *CaptainClient {
	this.observers = append(this.observers, observer)
	return this
}

func (this *CaptainClient) Online(name string) {
	oldstate := this.AllHealthy()
	this.watched[name] = true
	for _, observer := range this.observers {
		observer.Online(this, name)
	}
	if !oldstate && this.AllHealthy() {
		this.AllOnline()
	}
}

func (this *CaptainClient) Offline(name string) {
	this.watched[name] = false
	for _, observer := range this.observers {
		observer.Offline(this, name)
	}
}

func (this *CaptainClient) AllOnline() {
	for _, observer := range this.observers {
		observer.AllOnline(this)
	}
	waiter := this.waiter
	if waiter != nil {
		// non-blocking send
		select {
		case waiter <- true:
		default:
		}
	}
}

func (this *CaptainClient) IsHealthy(name string) bool {
	return this.watched[name]
}

func (this *CaptainClient) AllHealthy() bool {
	for _, state := range this.watched {
		if !state {
			return false
		}
	}
	return true
}

func (this *CaptainClient) WaitUntilAllOnline() *CaptainClient {
	this.waiter = make(chan bool, 1)
	return this
}

func (this *CaptainClient) Start() {
	go this.keeper.Start()
	if len(this.watched) == 0 {
		this.AllOnline()
	}
	if this.waiter != nil {
		<-this.waiter
		this.waiter = nil
	}
}

func (this *CaptainClient) Hang() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}

func (this *CaptainClient) Stop() {
	defer SilentOnPanic()
	this.CancelService()
	this.keeper.Quit()
}
