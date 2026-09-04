# 07 - Testing it

## What this shows

The same application as step 06, with tests. Two harnesses, for two different
kinds of claim.

- **`testing/unit`** runs one behavior with no node at all. It drives the
  callbacks directly and records everything the actor tried to do, so the
  assertions are about what it sent, called, registered or published.
- **`testing/stage`** starts real nodes and lets them talk over the real
  protocol. It is the only way to make a claim about two nodes.

Getting there needed one change to the code, and that is the most useful part of
this step.

## The change the tests forced

In step 06 the auditor reached for what it needed:

```go
a.Application().Behavior().(*taskrunnerApp).Record(m.ID, m.Took)
```

That line cannot run outside a node: there is no application, so
`Application()` has nothing to return and the type assertion panics. It is also
a hidden dependency and a type assertion that can fail at runtime.

Now the auditor asks for what it needs:

```go
type recorder interface {
	Record(id string, took time.Duration)
}
```

The application satisfies it and hands itself over in the member spec
(`Args: []any{a}`), the supervisor passes its args straight through, and the
auditor takes it in `Init`. A test passes a three-line fake instead. Nothing
about the running program changed - `results.log` is written by the same code -
but the auditor became a thing that can be run on its own.

That is the usual shape of this: the change that makes an actor testable is the
change that removes a hidden dependency.

## Run them

```bash
go test ./...
```

Under a second, including the two-node test.

## The unit tests

Six of them, in `apps/taskrunner/taskrunner_test.go`. The shape is always the
same: spawn, drive, assert.

```go
w, err := unit.Spawn(t, factoryWorker, gen.ProcessOptions{})
check.NoError(t, err)

w.SendMessage(from, messageSubmit{ID: "task-1", Effort: time.Millisecond})

rec, ok := w.ShouldSend().To(auditorName).Capture()
```

Three things are worth copying from these.

**Assert at the boundary, not on the fields.** The worker's job is to report,
so the test looks at what it sent. It checks `ID` and `For` and deliberately
does not check `Took`: that is a measurement, and a test that pins it down
fails on a busy machine for no reason.

**Assert the options, not just the behavior.** `ShouldRegisterEvent().Name(...)
.Notify(true).Buffer(10)` catches a wrong buffer size, which is otherwise
invisible until a subscriber arrives late and finds nothing waiting.

**Pin down the rules that surprised you.** Step 06 found that a pool forwards
everything in its Main queue, so a plain `Call` reaches a worker instead of the
pool. Two tests hold that in place: one that the high-priority request is
answered, one that the normal-priority request is forwarded and answers nothing.

```go
result, err := p.Call(gen.PID{}, requestAddWorkers{N: 2})
check.Nil(t, result)
p.ShouldForward().Once().Assert()
```

A test is only worth its line count if it fails when the code breaks. Change
`Buffer: 10` to `Buffer: 0` in the auditor and the harness prints what it saw
instead:

```
--- FAIL: TestAuditorRegistersWithNotifyAndBuffer
    check: expected exactly 1 check.RegisterEvent, got 0 match(es)
      observed:
        - RegisterEvent(pid=<...> name='tasks' notify=true buffer=0 open=false err=<nil>)
```

## The two-node test

`apps/taskrunner/stage_test.go` makes the claim the unit harness cannot: a task
encoded on the way out, an event encoded on the way back.

```go
s := stage.New(t)
runner := s.StartNode("runner", stage.NodeOptions{
	Applications: []gen.ApplicationBehavior{
		CreateApp(Options{Results: filepath.Join(t.TempDir(), "results.log")}),
	},
})
client := s.StartNode("client")
s.Connect(runner, client)
```

These are real nodes, so the application runs its real lifecycle - which is why
the results path had to become an option: a test writes to `t.TempDir()`, not
into the source tree.

Then a process on the client node subscribes and submits, and the assertion is
on the receiving side:

```go
client.ShouldReceiveEvent().To(pid).Once().Within(time.Second).Must()
```

`Within` is the difference between the two harnesses. The unit harness is
synchronous: by the time `SendMessage` returns, the callback has run. Real nodes
are not, so a stage assertion waits, and `Must` stops the test rather than
letting the rest run on a false premise.

## What the stage cannot catch

The nodes in a stage are separate nodes, but they are in one Go process, and the
EDF type registry is per process. Loading the application on the runner
registered the wire types for everything in the test binary, including the other
node.

So a node that forgets `RegisterTypes` fails in production and passes here. That
asymmetry is worth knowing before trusting a green stage test about wire format.

## Files

| File | What is in it |
|------|---------------|
| `apps/taskrunner/taskrunner_test.go` | The unit tests, and the fake recorder |
| `apps/taskrunner/stage_test.go` | Two real nodes |
| `apps/taskrunner/auditor.go` | Takes a recorder instead of reaching for one |
| `apps/taskrunner/app.go` | Hands itself over in `Args`, and takes a results path |
| `apps/taskrunner/supervisor.go` | Passes its args through to the auditor |

## Documentation

- [Testing Overview](https://docs.ergo.services/testing/overview) - which
  harness answers which kind of question
- [Unit](https://docs.ergo.services/testing/unit) - `Spawn`, the drivers, and
  every stub the subject can be given
- [Stage](https://docs.ergo.services/testing/stage) - real nodes, `Connect`,
  and what a stage can and cannot prove
- [Check](https://docs.ergo.services/testing/check) - the whole `Should*`
  grammar, `Within`, `Since`, `Capture` and `Collect`
- [Mock](https://docs.ergo.services/testing/mock) - the mock `gen`
  implementations both harnesses are built on

## Next

[08 - Looking inside](../08-observer) - the observer, and answering an AI agent's
questions about a running node.
