# 01 - The first actor

## What this shows

A worker that does tasks, and a client that gives them to it.

Three things you will use in everything after this:

- **`Send` is asynchronous.** The client hands over five tasks and is done in
  microseconds; the worker is still busy for two seconds. Nobody waits.
- **`Call` is synchronous.** When the client needs an answer - "how many did you
  handle?" - it blocks until the worker replies. That is the whole difference
  between `HandleMessage` and `HandleCall`.
- **State needs no lock.** The worker counts tasks in a plain struct field. One
  actor handles one message at a time on its own goroutine, so there is nothing
  to race with.

The worker is registered under a name, so the client addresses it as
`gen.Atom("worker")` rather than passing a PID around.

## Run it

```bash
go run .
```

It finishes in about two seconds and then keeps the node up so you can look at
it. Ctrl+C stops it.

## What you should see

```
08:45:40 [info] 3D8B778C: node 'tasks@localhost' built with "Ergo Framework:3.3.0" successfully started
08:45:40 [info] <...1071>: worker <...1071> is ready
08:45:40 [info] <...1072>: submitted task-1
08:45:40 [info] <...1072>: submitted task-2
...
08:45:41 [info] <...1071>: finished task-1 in 120ms
08:45:41 [info] <...1072>: got result for task-1 (1 of 5)
...
08:45:42 [info] <...1072>: all 5 tasks done, the worker was busy for 1.8s
08:45:42 [info] <...1072>: observer is at http://localhost:9911 - press Ctrl+C to stop
```

All five `submitted` lines appear before the first `finished` line. That is
`Send` not waiting.

Everything goes through the logger, and nothing here prints with `fmt`: the node
logs node lines, each actor logs its own, and the log carries the level, the
source and the time. The observer streams the same lines live, and a real
deployment routes them to a file or to Sentry without any of this code changing.

Then open **http://localhost:9911**. The worker and the client are in the
process list with their mailboxes and message counters; click the worker to see
what it is linked to and how long it has been running. Every step of this path
has the observer, so it is worth getting used to now.

Press Ctrl+C and look for the worker among the shutdown lines:

```
[info] <...1050>: worker stopped: <...0.1>: shutdown
```

That is `Terminate` running - the actor is told it is going away and can clean
up. The reason carries who asked: `<...0.1>` is the node core, because the node
itself is stopping. In step 02 the same line names the supervisor instead.

## Files

| File | What is in it |
|------|---------------|
| `types.go` | The messages. `Task`/`TaskDone` are sent, `StatsRequest`/`StatsResponse` are called |
| `worker.go` | The actor that does the work |
| `client.go` | The actor that submits tasks and collects results |
| `main.go` | Starts the node, spawns both, waits |

## Documentation

- [Actor Model](https://docs.ergo.services/basics/actor-model) - why an actor
  owns its state and why that removes the lock
- [Actor](https://docs.ergo.services/actors/actor) - every callback, the mailbox
  and its four queues
- [Process](https://docs.ergo.services/basics/process) - names, PIDs, and what
  `Send` and `Call` actually do
- [Handling Sync Requests](https://docs.ergo.services/advanced/handle-sync) -
  answering a `Call` later instead of holding the mailbox

## Next

[02 - When a worker dies](../02-supervision) - the worker crashes, and something
puts it back.
