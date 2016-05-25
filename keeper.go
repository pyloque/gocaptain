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
	started       bool
}

func NewServiceKeeper(client *CaptainClient) *ServiceKeeper {
	return &ServiceKeeper{client, 0, 10, 1000, make(chan bool), false}
}

func (this *ServiceKeeper) Start() {
	this.started = true
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
	if this.started {
		this.Stop <- true
	}
}
