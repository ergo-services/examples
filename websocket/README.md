## Project: "demo"

### Generated with
 - Types for the network messaging: true
 - Enabled Observer: true (http://localhost:9911)
 - Loggers: colored rotate
 

### Supervision Tree

Applications
 - `MyApp{}` ./examples/demo/apps/myapp/myapp.go
   - `MyActorInApp{}` ./examples/demo/apps/myapp/myactorinapp.go
   - `MySup{}` ./examples/demo/apps/myapp/mysup.go
     - `MyActorInSup{}` ./examples/demo/apps/myapp/myactorinsup.go
     - `MyTCP{}` ./examples/demo/apps/myapp/mytcp.go
     - `MyPool{}` ./examples/demo/apps/myapp/mypool.go
     - `MyWeb{}` ./examples/demo/apps/myapp/myweb.go
   - `MyUDP{}` ./examples/demo/apps/myapp/myudp.go

Messages are generated for the networking in ../../examples/demo/types.go
- `MyMsg1{}`
- `MyMsg2{}`


#### Used command

This project has been generated with the `ergo` tool. To install this tool, use the following command:

`$ go install ergo.services/tools/ergo@latest`

Below the command that was used to generate this project:

```$ ergo -path ./examples/ -init "demo{tls}" -with-app MyApp -with-actor MyApp:MyActorInApp -with-sup MyApp:MySup -with-actor "MySup:MyActorInSup{type:afo}" -with-tcp "MySup:MyTCP{port:12345,tls}" -with-udp "MyApp:MyUDP{port:2345}" -with-pool "MySup:MyPool{size:3}" -with-web "MySup:MyWeb{tls,websocket}" -with-msg MyMsg1 -with-msg MyMsg2 -with-observer -with-logger colored -with-logger rotate```
