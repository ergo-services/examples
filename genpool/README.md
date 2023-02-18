## Pool demo scenario ##

It starts a Pool process with the name "MyPool". This process spawns 5 worker processes which are handling messages/requests forwarded by the "MyPool" process.
There also "MyPing" process - it sends messages and makes a call requests to the "MyPool" process.

Here is output of this example
```
❯❯❯❯ go run .
Starting node: pool@localhost...OK
Starting mypool process ...
   started pool worker:  <AAF60CDC.0.1012>
   started pool worker:  <AAF60CDC.0.1013>
   started pool worker:  <AAF60CDC.0.1014>
   started pool worker:  <AAF60CDC.0.1015>
   started pool worker:  <AAF60CDC.0.1016>
OK
Starting myping process ...OK
MyPing send message 'Hello World'
[<AAF60CDC.0.1013>] received Info message: Hello World
MyPing cast message 'Hello World'
[<AAF60CDC.0.1014>] received Cast message: Hello World
MyPing make call request 'ping'
[<AAF60CDC.0.1015>] received Call request: ping
MyPing send message 'Hello World'
[<AAF60CDC.0.1016>] received Info message: Hello World
MyPing cast message 'Hello World'
[<AAF60CDC.0.1012>] received Cast message: Hello World


```
