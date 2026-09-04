# 03 - Making it an application

## What this shows

The supervisor from step 02, wrapped in an application. That wrapper is what a
real project ships, and it brings three things with it.

- **Somewhere to keep what the processes need.** The results file is opened in
  `Init`, before a single process exists, and closed in `Terminate`, after the
  last one is gone. The worker never opens it and never closes it - it reaches
  it through `w.Application().Behavior()`.
- **A group, and a rule about it.** `Group` lists what the application watches:
  here, the supervisor. The worker lives under the supervisor, so it is part of
  the application but not a member of the group. `Mode` decides what a member's
  death means. This one is `Transient`: a member that fails takes the whole
  application down with it, while a member that exits cleanly stops it only if
  it was the last one.
- **A package instead of a pile of process names.** `apps/taskrunner` exports
  four functions and nothing else. No caller knows that the worker is registered
  as `taskrunner_worker`, and no caller can build a message by hand.

## The lifecycle

`Load` only hands over the spec. The four phases after it run in this order, and
each one is worth a different kind of code:

| Phase | When | What belongs here |
|-------|------|-------------------|
| `Init` | before the group is spawned | open what the group will need |
| `Start` | once every member is up | announce, warm up, kick off |
| `Stop` | before the surviving members are asked to exit | stop taking new work |
| `Terminate` | after the last process of the application is gone | release what `Init` opened |

`Init` and `Terminate` are the pair that matters. Closing the file in `Stop`
would close it while workers are still writing to it.

## Run it

```bash
go run .
```

About three seconds of scenario, then the node stays up. Ctrl+C stops it.

Then look at what the application wrote:

```bash
cat results.log
```

## What you should see

Startup, in phase order:

```
[info] App#<...'taskrunner'>: init: opened results.log, no member is running yet
[info] <...1050>: supervisor <...1050> started
[info] <...1051>: worker <...1051> is ready
[info] App#<...'taskrunner'>: start: the group is up, mode is transient
```

Then three tasks, and a worker broken on purpose:

```
[info] <...1052>: the runner has handled 3 tasks
[info] <...1052>: --- breaking the worker, which is not a group member ---
[error] <...1051>: process terminated abnormally - simulated worker failure
[info] <...1053>: worker <...1053> is ready
[info] <...1052>: the runner answers again, and counts 1 - it started over
```

The application said nothing. A worker is not a member of the group, so its
death is the supervisor's business and no one else's.

Now the supervisor, which is a member:

```
[error] <...1050>: process terminated abnormally - simulated supervisor failure
[info] App#<...'taskrunner'>: stopping due to termination of <...1050> with reason: simulated supervisor failure
[info] App#<...'taskrunner'>: stop: going down (simulated supervisor failure)
[info] App#<...'taskrunner'>: stopped with reason simulated supervisor failure
[info] App#<...'taskrunner'>: terminate: the group is gone, closing results.log
[warning] <...1052>: the runner is gone: unknown process
[info] <...1052>: the application is loaded again and can be started once more, the node did not go down with it
```

That is `Transient` doing its job. The reason travels: the member's error becomes
the application's stop reason, and `Terminate` gets it too.

The node itself is fine. A stopped application goes back to `loaded`, which is
where it was before it started, and it can be started again.

## The other run

```bash
go run . -fail-init
```

`Init` returns an error, and two things follow:

```
[error] 3D8B778C: unable to start application 'taskrunner': no room on the disk for the results file
```

There is no `terminate:` line. **`Terminate` never runs if `Init` failed**, which
is exactly right: it would try to close a file that was never opened. Whatever
`Init` did before it gave up, it has to undo itself.

And the node is gone. An application listed in `NodeOptions.Applications` is not
optional - if it cannot start, the node stops. The `process terminated abnormally
- kill` lines after the error are that forced shutdown.

## Files

| File | What is in it |
|------|---------------|
| `apps/taskrunner/app.go` | The application: the spec, the four phases, the results file |
| `apps/taskrunner/api.go` | The four exported functions. This is the whole public surface |
| `apps/taskrunner/types.go` | The messages, all unexported |
| `apps/taskrunner/supervisor.go` | The group member |
| `apps/taskrunner/worker.go` | The worker, now writing through the application |
| `main.go` | Node options, and the flag for the failing run |
| `client.go` | Drives the scenario, using nothing but the package API |

The `apps/<name>/` directory is the usual layout: one directory per application,
and `main.go` at the top wiring them into a node.

## Documentation

- [Application](https://docs.ergo.services/basics/application) - the spec in
  full: the four phases, `Group`, `Mode`, `Depends`, `Env`, phase timeouts
- [Project Structure](https://docs.ergo.services/basics/project-structure) - the
  `apps/<name>/` layout, and the package API this step uses
- [Node](https://docs.ergo.services/basics/node) - `NodeOptions`, and what it
  means for an application to be listed there
- [Boilerplate Code Generation](https://docs.ergo.services/tools/ergo) - the
  `ergo` tool, which scaffolds exactly this shape

## Next

[04 - Two nodes](../04-network) - the same client, on another machine.
