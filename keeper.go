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
		this.client.ShuffleOrigin()
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
	flags := this.client.CheckDirty()
	if flags[0] {
		dirties := this.client.CheckServiceVersions()
		for _, name := range dirties {
			this.client.ReloadService(name)
		}
	}
	if flags[1] {
		dirties := this.client.CheckKvVersions()
		for _, key := range dirties {
			this.client.ReloadKv(key)
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
