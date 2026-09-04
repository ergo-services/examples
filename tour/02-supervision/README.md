# 02 - When a worker dies

## What this shows

The same worker as step 01, with a bug in it: it divides the effort over
`task.Items` without checking that there are any. A task with `Items: 0` panics
it.

Nobody catches that panic. Instead the worker is started by a **supervisor**,
and the supervisor puts it back. That is the whole idea: you do not defend every
actor against every mistake, you let it die and restart it clean.

Three things worth watching for:

- **The name survives, the PID does not.** The client always sends to
  `gen.Atom("worker")`. Behind that name a different process appears after each
  restart, with a new PID in the log.
- **A restart is a new instance.** The counter inside the worker goes back to
  zero. Whatever state the old one held is gone, which is why an actor that must
  not lose state has to keep it somewhere that outlives it.
- **Restarting is bounded.** The default is at most five restarts in any five
  seconds. Break it faster than that and the supervisor stops trying.

Everything in the supervisor here is the default: One For One, and the default
`Transient` restart policy, which restarts on an abnormal exit and leaves a
clean one alone. A panic is abnormal.

## Run it

```bash
go run .
```

About ten seconds. Ctrl+C stops it.

## What you should see

Act one, two ordinary tasks. Then one with no items:

```
[info]  <...1073>: --- one task with Items: 0 ---
[info]  <...1073>: submitted task-bad (items: 0)
[panic] <...1072>: Actor terminated. Panic reason: "integer divide by zero" at main.(*worker).HandleMessage[.../worker.go:30]
[error] <...1072>: process terminated abnormally - panic
[info]  <...1072>: worker <...1072> stopped: panic
[info]  <...1074>: worker <...1074> is ready
[info]  <...1073>: a worker answers again, tasks it has handled: 0
```

The panic line names your own file and line, not the runtime's - the framework
recovers the panic and reports where it came from, which is usually enough to
find it without a stack trace.

Note the PIDs: `1072` died, `1074` answers. Same name, different process, counter
back to zero.

Then one bad task every 400ms, and after the fifth the supervisor is done:

```
[error] <...1071>: process terminated abnormally - supervisor restart intensity exceeded (max 5 in 5s): limit exceeded: panic
[error] <...1073>: cannot submit bad-6: unknown process
[warning] <...1073>: nothing answers any more: unknown process
```

The supervisor terminated with `gen.ErrExceeded` wrapped together with the last
child failure, so a supervisor above it could tell what happened.

## Why one bad task at a time

Sending seven bad tasks at once produces exactly **one** crash, not seven: the
worker panics on the first, and everything still queued in its mailbox dies with
it. A crash takes the mailbox with it.

That is why act three sends one every 400ms - each has to land on a live worker
to cause its own crash. It is also a real lesson about retries: work in flight
when a process dies is gone unless something outside it remembers.

## Files

| File | What is in it |
|------|---------------|
| `supervisor.go` | The supervisor spec - it is short, and all defaults |
| `worker.go` | The same worker as step 01, plus the bug |
| `client.go` | Drives the three acts |
| `main.go` | Spawns the supervisor instead of the worker |

The only structural change from step 01 is in `main.go`: the node spawns the
supervisor, and the supervisor owns the worker.

## Documentation

- [Supervision Tree](https://docs.ergo.services/basics/supervision-tree) - the
  idea, and how a tree is shaped
- [Supervisor](https://docs.ergo.services/actors/supervisor) - the four types,
  restart strategies, intensity, significant children
- [Links And Monitors](https://docs.ergo.services/basics/links-and-monitors) -
  the exit signals a restart is built on
- [Actor Model Vet Tool](https://docs.ergo.services/tools/argus) - argus, which
  finds the callback mistakes the compiler cannot

## Next

[03 - Making it an application](../03-application) - what wraps the supervisor,
and who owns the things a worker needs.
