package gocaptain

import (
	"time"
)

type ServiceKeeper struct {
	client        *CaptainClient
	lastKeepTs    int64
	KeepAlive     int64
	CheckInterval time.Duration
	Stop          chan bool
}

func (this *ServiceKeeper) Start() {
	for {
		this.client.ShuffleUrlRoot()
		this.watch()
		this.keep()
		select {
		case <-this.Stop:
			break
		case <-time.After(this.CheckInterval * time.Millisecond):
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
	if now-this.lastKeepTs > this.KeepAlive {
		this.client.KeepService()
		this.lastKeepTs = now
	}
}

func (this *ServiceKeeper) Quit() {
	this.Stop <- true
}