# 04 - Two nodes

## What this shows

The same application as step 03, on its own node, and a client on another node
that uses it through the same package API. The only thing that changed in the
call is that it now names the node to send to.

- **The name is the address.** `runner@localhost` is all the client is given.
  Nothing configures a host, a port or a connection.
- **The types have to cross.** A message between nodes is encoded, so its type
  must be registered on both sides before the first send. Every field must be
  exported; the type itself need not be.
- **Failure is visible, and not by timeout.** The client monitors the node. When
  it goes away the monitor says so at once, and sends to it fail immediately with
  `no route` instead of hanging.

## Run it

Two terminals, in either order:

```bash
go run ./cmd/runner
```

```bash
go run ./cmd/client
```

Start the client first and it simply says `no route` once a second until the
runner appears. Nothing has to be started in a particular order.

Nodes on one machine find each other through a registrar that the first node to
start hosts on port 4499. That is a development convenience - a real cluster
uses etcd or Saturn - but it means two `go run` commands are the whole setup.

## What you should see

On the runner, a connection and then work arriving from somewhere else:

```
[info] BC68B0F0: new connection with 'client@localhost' (8FA01FD5)
[info] <...1051>: finished task-1 for 'client@localhost' in 150ms
```

`HandleMessage` did not have to change for any of this. A remote send arrives
exactly like a local one, and `from` names the sender, node included.

On the client, the tasks going out and the answers coming back:

```
[info] <...1050>: talking to the runner on 'runner@localhost'
[info] <...1050>: submitted task-1
[info] <...1050>: watching 'runner@localhost'
[info] <...1050>: the runner has handled 3 tasks
```

The `watching` line comes after the first send, not in `Init`. A monitor needs a
live connection: `MonitorNode` on a node nothing has reached yet fails with
`no route`, so the client arms it once a send has gone through.

## Kill the runner

Ctrl+C the runner terminal and watch the client:

```
[info]    8FA01FD5: connection with 'runner@localhost' (BC68B0F0) terminated with reason: normal
[warning] <...1050>: the node 'runner@localhost' is down - the monitor said so, no timeout was needed
[warning] <...1050>: cannot submit task-36: no route
[warning] <...1050>: cannot read stats: no route
```

Two different things happened there. `MonitorNode` delivered a
`gen.MessageDownNode` the moment the connection dropped - that is the fast path,
and it is why a client does not have to discover the loss by waiting for a `Call`
to time out. And every send after that failed straight away with `no route`,
rather than queueing somewhere in the hope that the node returns.

A monitor is spent once it fires. The client clears its flag and arms a new one
on the next successful send, so a node that comes and goes is reported every
time, not only the first.

## Bring it back

Start the runner again, in the same terminal, and change nothing on the client:

```
[warning] <...1050>: cannot submit task-49: no route
[info]    8FA01FD5: new connection with 'runner@localhost' (BC68B0F0)
[info]    <...1050>: submitted task-50
[info]    <...1050>: watching 'runner@localhost'
[info]    <...1050>: the runner has handled 2 tasks
```

There is no reconnection code in this example. The next send resolved the name
again, and the connection came back with it. Note the count: the new runner is a
new process with a fresh counter, exactly as in step 02, only now the thing that
restarted is a whole node.

## Registering the types

Two nodes, two ways to arrive at the same registration:

- The runner hosts the application, and the spec carries the list:
  `Network: gen.ApplicationNetwork{RegisterTypes: wireTypes}`. Loading the
  application registers them.
- The client hosts nothing, so `cmd/client/main.go` calls
  `taskrunner.RegisterTypes(node)` right after the node starts.

Both have to happen before the first message goes out. Registration is a
node-level fact, not a per-message one. Take the client's call away and every
send says so - no panic, no silent drop:

```
[warning] <...1051>: cannot submit task-1: no encoder for type taskrunner.messageSubmit
```

If a field of a wire type is unexported, registration refuses the type with
`struct ... has unexported field(s)`. That is deliberate: the other node cannot
be given what its own package would not let it read.

One caveat about the client's way. It registers after the node has started, so
for a moment the node is already accepting connections while its type table is
still empty. Nothing dials into the client here, so that window is empty - but a
node others call into should get its types from an application spec, the way the
runner does, and not from a call in `main`.

## If the nodes will not talk

```
[warning] unable to handshake with 127.0.0.1:54023: incorrect digest (accept stage 'hello')
```

Different cookies. Both nodes here set `options.Network.Cookie = "taskrunner"`,
and the handshake fails if they disagree. You will also see this line if some
other Ergo node happens to be running on the same machine: it finds the new node
through the shared registrar, tries to connect, and is refused. That is the
cookie doing its job.

## Files

| File | What is in it |
|------|---------------|
| `cmd/runner/main.go` | The node that hosts the application |
| `cmd/client/main.go` | The node that uses it, and the type registration |
| `cmd/client/client.go` | The client actor: the monitor, and one task a second |
| `apps/taskrunner/api.go` | The API, now with a node to send to |
| `apps/taskrunner/types.go` | The wire types |
| `apps/taskrunner/app.go` | The spec, now declaring what it puts on the wire |

Each node has its own observer: http://localhost:9911 for the runner,
http://localhost:9912 for the client.

## Documentation

- [Network Transparency](https://docs.ergo.services/networking/network-transparency) -
  addressing another node, and registering the types that cross
- [Network Stack](https://docs.ergo.services/networking/network-stack) - the
  handshake, the cookie, flags, message size and fragmentation
- [Service Discovering](https://docs.ergo.services/networking/service-discovering) -
  the registrar, and why the embedded one is for development
- [Links And Monitors](https://docs.ergo.services/basics/links-and-monitors) -
  `MonitorNode` among the rest
- [Message Versioning](https://docs.ergo.services/advanced/message-versioning) -
  changing a wire type without breaking the node that has not been redeployed
- [Building a Cluster](https://docs.ergo.services/advanced/building-a-cluster) -
  what replaces two `go run` commands in production

## Next

[05 - Events](../05-events) - one producer, many listeners, and nobody wiring
them together.
