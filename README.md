Captain
--------------------------
Captain is yet another service discovery implementation based on redis.
Captain sacrifices a little high availability for simplicity and performance.
In most cases, we dont have so many machines as google/amazon.
The possibility of machine crashing is very low, high Availability is not so abviously important yet.
But the market only provides zookeeper/etcd/consul, they are complex, at least much complexer compared with captain.

Use Golang Captain Client
-------------------------------
```go
import "github.com/pyloque/gocaptain"

type Observer struct {
	name string
}

func (this *Observer) Online(name string) {
	println(name + " is ready")
}
func (this *Observer) AllOnline() {
	println(this.name + " is all ready")
    println(client.Select("service1").UrlRoot()) // now select the service your want
    println(client.Select("service2").UrlRoot())
}
func (this *Observer) Offline(name string) {
	println(name + " is offline")
}

var client *gocaptain.CaptainClient

func main() {
    // connect multiple captain servers
    client = gocaptain.NewCaptainClientWithOrigins(
        gocaptain.NewServiceItem("localhost", 6789),
        gocaptain.NewServiceItem("localhost", 6790))
	// client := gocaptain.NewCaptainClient("localhost", 6789) // connect single captain server
	client.Watch("service1", "service2", "service3")  // define service dependencies
          .Provide("service4", gocaptain.NewServiceItemWithTtl("localhost", 6400, 30))  // provide service with ttl of 30s
          .Observe(&Observer{"service"}) // observe status change of service dependencies
          .KeepAlive(10) // keepalive heartbeat in seconds for provided service
          .CheckInterval(1000) // check service dependencies with 1000ms interval
          .Start()
	client.Hang() // hang just for test
	client.Stop() // cancel provided service
}
```
