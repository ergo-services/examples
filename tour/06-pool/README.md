# 06 - A pool of workers

## What this shows

One task at a time was fine while there was one client sending one task a
second. Here the client sends sixty at once, and the single worker becomes a
pool of them behind the same registered name.

- **The caller does not change.** `taskrunner.Submit` is the same line of code as
  in step 05. Behind the name is now `act.Pool`, which hands each message to one
  of its workers in turn.
- **The pool is not a queue.** Each worker has a bounded mailbox. When all of
  them are full, the pool drops the message and says so. Nothing is buffered
  anywhere on the caller's behalf.
- **`Send` returning nil means nothing about the outcome.** All sixty sends
  succeed. Twenty-seven of those tasks are never done.

## Run it

Two terminals, in either order:

```bash
go run ./cmd/runner
```

```bash
go run ./cmd/client
```

The client runs two rounds of sixty tasks, about twenty seconds in total.

## What you should see

Round one, three workers:

```
[info]    <...1050>: --- round 1: 60 tasks as fast as they will go ---
[info]    <...1050>: 60 submitted, every Send returned without an error
[info]    <...1050>: round 1: 60 submitted, 33 finished
[warning] <...1050>: 27 tasks were dropped - the pool is not a queue
```

Thirty-three is not a random number. The pool has three workers with a mailbox
of ten each, and each one is busy with a task while its mailbox fills: three in
hand plus thirty queued. Task thirty-four arrives, the pool tries every worker,
every mailbox is full, and on the runner side you get one line per dropped
message:

```
[error] <...1054>: no available worker process. ignored message from <...1050>
```

Twenty-seven of those, matching the client's count exactly.

Then the pool grows and the same flood goes through untouched:

```
[info] <...1050>: --- pool grown to 8 workers, same flood again ---
[info] <...1050>: round 2: 60 submitted, 60 finished
```

Eight workers with ten slots each is eighty-eight, comfortably more than sixty.
On the runner side, round two is done by eight different PIDs.

## Reaching the pool instead of a worker

`AddWorkers` looks like an ordinary `Call`, but it is not:

```go
process.CallWithPriority(target(runner), requestAddWorkers{N: n}, gen.MessagePriorityHigh)
```

The pool forwards its whole Main queue to workers without inspecting it, and a
request at normal priority is in that queue like anything else - it would land
on a worker, which knows nothing about the pool it belongs to. High priority
goes to the System queue, which the pool handles itself. That is the whole
mechanism for talking *to* a pool rather than *through* it.

## Why the event moved

In step 05 the worker owned the `tasks` event. That cannot survive a pool: an
event name has exactly one owner, and there are several workers now. So the
auditor registers it, the workers report to the auditor, and the auditor both
writes the file and publishes.

Nobody outside notices. An event belongs to a node, not to a PID, and the client
subscribes to exactly the same `taskrunner.Tasks(runner)` as before. The change
is an improvement for the subscriber, in fact: a worker coming and going no
longer costs anyone a subscription.

One thing that was invisible in step 05 shows up now. The only subscriber is on
the other node, so stopping the client takes the count back to zero:

```
[info] <...1053>: 'tasks' has listeners now
[info] <...1053>: 'tasks' has no listeners left
```

## What the counters say

The pool answers `gen.Inspect` with its own numbers - `ergo:pool_size`,
`ergo:messages_forwarded`, `ergo:messages_unhandled`, `ergo:worker_restarts`.
Open the observer at http://localhost:9911, find `taskrunner_worker` in the
process list, and `messages_unhandled` is the twenty-seven. That number is the
one to watch in a real deployment: it is the only place a dropped message is
recorded.

## What to do about the drops

Not a bigger mailbox. A mailbox of a thousand would have accepted the flood and
then taken fifty seconds to work through it, with the client none the wiser.

The real options are the ones the pool makes visible:

- **More workers**, which is what round two does. Sensible when the work is
  waiting on something other than a CPU.
- **Backpressure**, meaning a `Call` instead of a `Send`. A caller that waits for
  each task cannot outrun the pool by construction.
- **Somewhere durable to put the work**, if neither of those is honest - a queue
  outside the node, so that a dropped message is a retry rather than a loss.

The pool also replaces workers that die: if forwarding finds a worker gone, it
spawns a new one and counts it in `worker_restarts`. Nothing supervises the
workers individually, and nothing has to.

## Files

| File | What is in it |
|------|---------------|
| `apps/taskrunner/pool.go` | The pool: size, worker mailbox, and the one high-priority request |
| `apps/taskrunner/worker.go` | The worker, now anonymous and one of several |
| `apps/taskrunner/auditor.go` | Owns the event now, and still the only writer |
| `apps/taskrunner/api.go` | `Submit` unchanged, `AddWorkers` added |
| `cmd/client/client.go` | Two rounds of sixty, counting what actually finished |

## Documentation

- [Pool](https://docs.ergo.services/actors/pool) - the options, `AddWorkers` and
  `RemoveWorkers`, and the rule that only the Main queue is forwarded
- [Actor](https://docs.ergo.services/actors/actor) - the four mailbox queues the
  priorities map onto
- [Router](https://docs.ergo.services/actors/router) - when the choice of worker
  has to depend on the message rather than on turn
- [Inspecting Actor State](https://docs.ergo.services/advanced/inspecting-state) -
  `HandleInspect`, where the pool's counters come from

## Next

[07 - Testing it](../07-testing) - the same actors, without a node.
