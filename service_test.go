package gocaptain

import (
	"testing"
)

func TestService(t *testing.T) {
	locals := LocalService{
		1,
		map[string]int64{"watch": 2},
		map[string][]*ServiceItem{"watch": []*ServiceItem{NewServiceItem("localhost", 6000)}},
		map[string][]*ServiceItem{},
	}
	if locals.GetVersion("watch") != 2 {
		t.Error("Get Version return illegal version")
	}
	if locals.GetVersion("watch1") != -1 {
		t.Error("Get Version return illegal version")
	}
	item := locals.RandomService("watch")
	if item.Host != "localhost" || item.Port != 6000 {
		t.Error("RandomService return illegal service item")
	}
	func() {
		defer func() {
			if err := recover(); err == nil {
				t.Error("RandomService should panic error")
			}
		}()
		locals.RandomService("watch1")
	}()
	locals.SetVersion("watch", 3)
	if locals.GetVersion("watch") != 3 {
		t.Error("Get Version return illegal version")
	}
	locals.ReplaceService("watch", []*ServiceItem{NewServiceItem("localhost", 6100)})
	item = locals.RandomService("watch")
	if item.Host != "localhost" || item.Port != 6100 {
		t.Error("RandomService return illegal service item")
	}
}
