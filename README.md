Captain
--------------------------
Captain is yet another service discovery implementation based on redis.
Captain sacrifices a little high availability for simplicity and performance.
In most cases, we dont have so many machines as google/amazon.
The possibility of machine crashing is very low, high Availability is not so abviously important yet.
But the market only provides zookeeper/etcd/consul, they are complex, at least much complexer compared with captain.
https://github.com/pyloque/captain

Use Golang Captain Client
-------------------------------
```go

import "fmt"
import "github.com/pyloque/gocaptain"

type Observer struct {
	name string
}

func (this *Observer) Online(client *gocaptain.CaptainClient, name string) {
	println(name + " is ready")
}
func (this *Observer) AllOnline(client *gocaptain.CaptainClient) {
	println(this.name + " is all ready")
    println(client.Select("service1").UrlRoot()) // now select the service your want
    println(client.Select("service2").UrlRoot())
}
func (this *Observer) Offline(client *gocaptain.CaptainClient, name string) {
	println(name + " is offline")
}

func (this *Observer) KvUpdate(client *gocaptain.CaptainClient, key string) {
    fmt.Printf("%v:%v", key, client.GetKv(key))
}


func main() {
    // connect multiple captain servers
    client := gocaptain.NewCaptainClientWithOrigins(
        gocaptain.NewServiceItem("localhost", 6789),
        gocaptain.NewServiceItem("localhost", 6790))
	// client := gocaptain.NewCaptainClient("localhost", 6789) // connect single captain server
	client.Watch("service1", "service2", "service3")  // define service dependencies
          .Failover("service1", NewServiceItem("localhost", 6100))  // provided failover services
          .Provide("service4", gocaptain.NewServiceItemWithTtl("localhost", 6400, 30))  // provide service with ttl of 30s
          .Observe(&Observer{"service"}) // observe status change of service dependencies
          .WatchKv("project_settings_service1")
          .KeepAlive(10) // keepalive heartbeat in seconds for provided service
          .CheckInterval(1000) // check service dependencies with 1000ms interval
          .WaitUntilAllOnline() // let Start method block until all dependent services are ready
          .Start()
	client.Hang() // hang just for test
	client.Stop() // cancel provided service
}
```
