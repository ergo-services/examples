# 10 - An external program

## What this shows

Not everything is going to be an actor. Some of the work you need is already
written, in another language, by somebody else - a solver, a converter, a report
generator, a tool from 1998 that has never been wrong. This step runs one of
those and talks to it.

Here the program is `awk`: it reads one line per finished task and prints a
running summary, and an actor feeds it and logs what comes back. Nothing about
the rest of the tour changes - the pool, the auditor and the event carry on
exactly as in step 08, and the new process simply subscribes to the same event.

## How you run a program

Three things, and that is the whole of it:

```go
port, err := meta.CreatePort(meta.PortOptions{
	Cmd:  "awk",
	Args: []string{"-f", "report.awk"},
	Tag:  "report",
})
alias, err := r.SpawnMeta(port, gen.MetaOptions{})
```

From then on:

- what you **send** to `alias` goes to the program's **stdin**
- what the program prints on **stdout** arrives as a message to your actor
- so does what it writes to **stderr**, and so do its start and its death

In text mode, which is the default, both directions use the same type:

```go
// writing
r.Send(r.port, meta.MessagePortText{Text: "task-1 150\n"})

// reading, in HandleMessage
case meta.MessagePortText:
	r.Log().Info("report: %s", strings.TrimRight(m.Text, "\r\n"))
```

There is no `exec.Command` in this example, no goroutine reading a pipe, no
buffer to manage. The reading happens off your actor's goroutine, and what
arrives is an ordinary message in an ordinary mailbox.

## Run it

```bash
go run ./cmd/runner
```

```bash
go run ./cmd/client
```

`awk` is already on any Unix machine, so there is nothing to install. Run the
runner from the module root: the program is started as `awk -f report.awk`, and
that path is relative to the working directory.

## What you should see

```
[info] <...1058>: reporter <...1058> is ready, feeding awk through Alias#<...>
[info] <...1058>: awk started, port Alias#<...>
[info] <...1058>: report: 1 done, average 150.0 ms
[info] <...1058>: report: 2 done, average 150.0 ms
```

The counting and the average happen in `report.awk`, three lines of it. The
actor never adds anything up.

## Kill the program

The interesting part. Find the `awk` process and kill it:

```bash
pkill -f "awk -f report.awk"
```

The log tells the whole story in four lines:

```
[info]  <...1058>: awk is gone, port Alias#<...>
[error] <...1058>: process terminated abnormally - unknown process
[info]  <...1058>: reporter <...1058> stopped: unknown process
[info]  <...1075>: reporter <...1075> is ready, feeding awk through Alias#<...>
```

Read it backwards. The program died, so the port terminated and said so. The
next task's line then had nowhere to go: sending to a dead alias returns
`gen.ErrProcessUnknown`, the reporter returns that error from its callback, and
it dies too - deliberately, because a reporter with no program is not worth
keeping. The supervisor restarts it, `Init` starts a fresh `awk`, and the events
buffered while all this happened are replayed into it.

Note what the new `awk` counts from: one. The program's state died with the
program. The framework can hand back the events it buffered; it cannot hand back
what a child process held in memory.

## While we are here: meta-processes

That `SpawnMeta` is the framework's way of putting blocking I/O next to an actor
without breaking it. The I/O runs on its own goroutine and reaches your actor as
messages, so the actor keeps handling one message at a time and never blocks in
a callback.

Mostly you do not write one. The framework ships the ones you are likely to
need, and a port is just the one that happens to fit "an external program":

| | |
|---|---|
| [Port](https://docs.ergo.services/meta-processes/port) | an external program over stdin/stdout, text or binary, with framing |
| [TCP](https://docs.ergo.services/meta-processes/tcp) | a TCP server and its connections |
| [UDP](https://docs.ergo.services/meta-processes/udp) | a UDP server |
| [Web](https://docs.ergo.services/meta-processes/web) | an HTTP server and its handlers |
| [WebSocket](https://docs.ergo.services/extra-library/meta-processes/websocket) | WebSocket handlers, in `ergo.services/meta/websocket` |
| [SSE](https://docs.ergo.services/extra-library/meta-processes/sse) | server-sent events, in `ergo.services/meta/sse` |

They all follow the shape you have just seen: create it, `SpawnMeta` it, get
messages. Writing your own is a last resort, and
[Meta-Process](https://docs.ergo.services/basics/meta-process) is where that is
described.

## Files

| File | What is in it |
|------|---------------|
| `apps/taskrunner/reporter.go` | The actor that owns the port and reads its output |
| `report.awk` | The external program. Three lines, one of which is `fflush` |
| `apps/taskrunner/supervisor.go` | A third child, last in dependency order |

## The one thing that will waste your afternoon

`fflush()` in `report.awk`. Without it `awk` buffers its output, the buffer is
never full, and the actor sits there receiving nothing while the program is
working perfectly. Every language does this on a pipe: Python needs
`flush=True`, C needs `fflush(stdout)`, Go's `bufio.Writer` needs `Flush`. When
a port goes quiet, suspect the program's buffering before anything in the
framework.

## Documentation

- [Port](https://docs.ergo.services/meta-processes/port) - text and binary mode,
  chunking and framing, buffer pools, environment, routing, inspection
- [Meta-Process](https://docs.ergo.services/basics/meta-process) - what a
  meta-process is, and what it is allowed to do
- [Links And Monitors](https://docs.ergo.services/basics/links-and-monitors) -
  watching a meta-process rather than dying with it

## That is the tour

Ten steps: an actor, a supervisor, an application, a second node, events, a
pool, tests, the tools for looking at all of it, an Erlang node driving the same
code, and an external program doing part of the work. The framework has more -
cron, tracing, TCP and web servers, TLS, a registrar for real clusters - and the
reference for those is https://docs.ergo.services.
