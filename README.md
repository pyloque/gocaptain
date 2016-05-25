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
}
func (this *Observer) Offline(name string) {
	println(name + " is offline")
}

func main() {
    // connect multiple captain servers
    client := gocaptain.NewCaptainClientWithOrigins(
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
