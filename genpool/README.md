## Pool demo scenario ##

It starts a Pool process with the name "mypool". This process spawns 5 worker processes that handle messages/requests forwarded by the "mypool" process. There is also "myping" process - it sends messages and makes call requests to the "mypool" process.

![image](https://user-images.githubusercontent.com/118860/221194341-628939e0-7be2-41bf-9f54-5374ac802a69.png)

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
myping send message 'Hello World'
[<AAF60CDC.0.1013>] received Info message: Hello World
myping send cast message 'Hello World'
[<AAF60CDC.0.1014>] received Cast message: Hello World
myping make a call request 'ping'
[<AAF60CDC.0.1015>] received Call request: ping
myping send message 'Hello World'
[<AAF60CDC.0.1016>] received Info message: Hello World
myping cast message 'Hello World'
[<AAF60CDC.0.1012>] received Cast message: Hello World

```
