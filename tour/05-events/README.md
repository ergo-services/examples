# 05 - Events

## What this shows

In step 04 the client asked, every few seconds, how many tasks were done. Here
it is told. The worker publishes one event per finished task and does not know,
or care, who reads it.

- **One publish, any number of readers.** The worker calls `SendEvent` once.
  Two processes get it: an auditor on the same node, and the client on the other
  node. Adding a third changes nothing in the worker.
- **A subscription comes with a backlog.** The event has a buffer, and
  subscribing returns whatever is in it. Throwing away that return value is the
  standard way to lose events.
- **Two ways to subscribe.** `MonitorEvent` says "tell me if it goes away".
  `LinkEvent` says "then I have no reason to live". The auditor uses the second
  one and the client the first, and the difference is visible.

The application also has a second process now. The auditor is the only thing
that writes `results.log`, driven by the event rather than by the worker. That
matters in the next step: however many workers there are, the file keeps one
writer.

## Run it

Two terminals, in either order:

```bash
go run ./cmd/runner
```

```bash
go run ./cmd/client
```

## What you should see

On the runner, the group is two processes now, and the worker notices that
somebody is listening:

```
[info] <...1050>: supervisor <...1050> started
[info] <...1051>: worker <...1051> is ready, publishing 'tasks'
[info] <...1052>: auditor <...1052> is ready, 0 event(s) were waiting
[info] <...1051>: 'tasks' has listeners now
[info] <...1051>: finished task-1 for 'client@localhost' in 150ms
```

On the client, no polling anywhere:

```
[info] <...1050>: submitted task-1
[info] <...1050>: subscribed to the runner, 0 event(s) were waiting
[info] <...1050>: task-1 is done in 150ms (1 of 1 submitted)
```

`HandleEvent` is where a subscription lands. Note what is not in it: no PID, no
process name, nothing about who published. The client subscribed to a name on a
node, and that is the whole coupling.

## The buffer

Leave the runner going and restart the client. It arrives to a stream that has
been running without it:

```
[info] <...1050>: subscribed to the runner, 6 event(s) were waiting
[info] <...1050>: already done before we arrived: task-1 in 150ms
[info] <...1050>: already done before we arrived: task-2 in 150ms
...
```

Those six came from `MonitorEvent`, not from `HandleEvent`. The producer's
buffer holds the last ten (`gen.EventOptions{Buffer: 10}`), and a subscriber
that ignores the returned slice silently starts from the present.

## Link or monitor

The auditor cannot do its job without the event, so it uses `LinkEvent` and dies
with it. The client has other work, so it uses `MonitorEvent` and gets a
`gen.MessageDownEvent` instead.

That is also why the supervisor is `Rest For One` here rather than `One For One`.
The children are declared in dependency order - worker, then auditor - so a
worker that dies takes the auditor with it and both come back in that order, the
event registered again before anything tries to subscribe. An auditor that dies
alone is restarted alone, because nothing depends on it.

## When the producer goes away

Ctrl+C the runner:

```
[warning] <...1050>: the event 'tasks' is gone: shutdown
[warning] <...1050>: cannot submit task-5: no route
```

The subscription is spent, exactly like the node monitor in step 04, and the
client clears its flag. Start the runner again and the next successful send
subscribes to the new worker's event:

```
[info] <...1050>: subscribed to the runner, 0 event(s) were waiting
[info] <...1050>: task-16 is done in 150ms (5 of 16 submitted)
```

Zero waiting, because that worker is new and has published nothing yet.

## About "has listeners now"

`gen.EventOptions{Notify: true}` is what makes the worker hear about its own
subscribers. Two things are worth knowing:

- The messages fire when the count **leaves zero** and when it **returns to
  zero**, not once per subscriber. That is why the line appears when the auditor
  subscribes and not again when the client does.
- In this example the auditor never unsubscribes, so the count never gets back
  to zero and `MessageEventStop` stays quiet. It is handled anyway: a producer
  that asks to be notified and then ignores half the notifications is a bug
  waiting for the day the count does reach zero.

The point of `Notify` is that a producer can stop working when nobody is
listening - polling a database, collecting statistics, keeping a connection open.

## Files

| File | What is in it |
|------|---------------|
| `apps/taskrunner/events.go` | The event name and its payload - the only exported type in the package |
| `apps/taskrunner/worker.go` | Registers the event, publishes, and hears about listeners |
| `apps/taskrunner/auditor.go` | Links to the event and writes the results file |
| `apps/taskrunner/supervisor.go` | Two children, in dependency order |
| `cmd/client/client.go` | Subscribes and handles the backlog |

## Documentation

- [Events](https://docs.ergo.services/basics/events) - registering, publishing,
  subscribing, and what `Notify` and `Buffer` do
- [Pub/Sub Internals](https://docs.ergo.services/advanced/pub-sub-internals) -
  what one publish costs when the subscribers are on other nodes
- [Links And Monitors](https://docs.ergo.services/basics/links-and-monitors) -
  `LinkEvent` against `MonitorEvent`, and the rest of the pairs
- [Supervisor](https://docs.ergo.services/actors/supervisor) - Rest For One, and
  the other three types

## Next

[06 - A pool of workers](../06-pool) - one name, several workers behind it, and
what happens when they cannot keep up.
