package gocaptain

import (
	//	"log"
	//	"runtime"
	"time"
)

type ServiceKeeper struct {
	client     *CaptainClient
	lastKeepTs int64
	Ttl        int64
	Stop       chan bool
}

func (this *ServiceKeeper) Start() {
	for {
		this.client.ShuffleUrlRoot()
		this.watch()
		this.keep()
		select {
		case <-this.Stop:
			break
		case <-time.After(time.Second):
		}
	}
}

func (this *ServiceKeeper) watch() {
	defer SilentOnPanic()
	if this.client.CheckDirty() {
		dirties := this.client.CheckVersions()
		for _, name := range dirties {
			this.client.ReloadService(name)
		}
	}
}

func (this *ServiceKeeper) keep() {
	defer SilentOnPanic()
	now := time.Now().Unix()
	if now-this.lastKeepTs > this.Ttl {
		this.client.KeepService()
		this.lastKeepTs = now
	}
}

func (this *ServiceKeeper) Quit() {
	this.Stop <- true
}
