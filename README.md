Use Golang Captain Client
-------------------------------
```
package main

import "github.com/pyloque/gocaptain"

type Observer struct {
	name string
}

func (this *Observer) Ready(name string) {
	println(name + " is ready")
}
func (this *Observer) AllReady() {
	println(this.name + " is all ready")
}
func (this *Observer) Offline(name string) {
	println(name + " is offline")
}

func main() {
	client1 := gocaptain.NewCaptainClient("localhost", 6789)
	client1.Provide("service1", gocaptain.NewServiceItem("localhost", 6100)).Observe(&Observer{"service1"}).Start()
	client2 := gocaptain.NewCaptainClient("localhost", 6789)
	client2.Provide("service2", gocaptain.NewServiceItem("localhost", 6200)).Observe(&Observer{"service2"}).Start()
	client3 := gocaptain.NewCaptainClient("localhost", 6789)
	client3.Watch("service1", "service2").Provide("service3", gocaptain.NewServiceItem("localhost", 6300)).Observe(&Observer{"service3"}).Start()
	client4 := gocaptain.NewCaptainClient("localhost", 6789)
	client4.Watch("service1", "service2", "service3").Provide("service4", gocaptain.NewServiceItem("localhost", 6400)).Observe(&Observer{"service4"}).Start()
	client4.Hang()
	client1.Stop()
	client2.Stop()
	client3.Stop()
	client4.Stop()
}
```
