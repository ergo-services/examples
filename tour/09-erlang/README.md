# 09 - Talking to Erlang

## What this shows

The same task runner, driven from an Erlang node. An `erl` shell calls
`gen_server:call`, the task goes through the pool, the auditor writes it down.

Two things are worth taking from this step.

- **The node changes, the code does not.** Three lines in `main.go` replace the
  network stack: EPMD instead of the Ergo registrar, and the Erlang handshake
  and DIST protocol instead of the Ergo ones. Nothing in `apps/taskrunner`
  differs from step 08 by a single character.
- **The boundary is one process.** A single actor embeds
  `erlang23.GenServer`, which is what makes `gen_server:call` and
  `gen_server:cast` arrive as callbacks. It translates what Erlang sends into
  the package API from step 03, and the pool, the auditor and the event never
  learn that Erlang exists.

## Read the type mapping first

This is where the mistakes come from, and they are silent ones. Erlang sends
ETF terms, and the decoder converts them to the Go types it has - not the ones
you would have guessed:

| Erlang | Go |
|--------|-----|
| integer | `int64` |
| binary `<<"task-1">>` | `[]byte` |
| atom | `gen.Atom` |
| tuple | `etf.Tuple` |
| list | `etf.List` |
| string `"abc"` | `etf.List` of numbers - use `etf.TermToString` |
| map | `map[any]any` |

The full table, both directions, is in
[Erlang](https://docs.ergo.services/extra-library/network-protocols/erlang#etf-data-format)
documentation. **Read it before writing a callback.** Writing `.(int)` where the
term is an `int64` costs an afternoon: the type assertion just fails, the
callback takes its error branch, and Erlang gets a perfectly well-formed
`{error, badarg}` with nothing in any log to say why. That exact mistake is in
the history of this example.

The `%T` in the two warnings in `gate.go` is there for the same reason: when a
term is not what the code expects, the log says what it actually was.

## Run it

```bash
go run ./cmd/runner
```

It waits, and prints the two commands to type. In another terminal:

```bash
erl -sname erl@localhost -setcookie taskrunner
```

Then, in the Erlang shell:

```erlang
gen_server:call({taskrunner_gate, 'runner@localhost'}, {submit, <<"task-1">>, 150}).
```

Erlang must be installed for this step - `brew install erlang`, or your
distribution's package. Nothing else: the Ergo node runs its own EPMD, so there
is no daemon to start.

## What you should see

Erlang answers:

```erlang
ok
```

And the Ergo side, one task from beginning to end:

```
[info] BC68B0F0: new connection with 'erl@localhost' (E8B15506)
[info] <...1045>: submitted task-1 for 'erl@localhost'
[info] <...1042>: finished task-1 in 150ms
```

`cat results.log` shows the line the auditor wrote. The path was: Erlang ->
`gate.HandleCall` -> `taskrunner.Submit` -> the pool -> a worker -> the auditor
-> the file. Only the first hop knew anything about Erlang.

The other two directions work as well:

```erlang
gen_server:cast({taskrunner_gate, 'runner@localhost'}, {hello, from_erlang}).
erlang:send({taskrunner_gate, 'runner@localhost'}, just_a_message).
```

arriving as `HandleCast` and `HandleInfo`:

```
[info] <...1045>: cast: etf.Tuple{"hello", "from_erlang"}
[info] <...1045>: info: "just_a_message"
```

Both print through `%#v`, which shows a `gen.Atom` as a bare quoted string. They
are atoms, not strings - see the table above.

## Two things that will confuse you

**`net_adm:ping` returns `pang` even though everything works.**

```erlang
1> net_adm:ping('runner@localhost').
pang
2> gen_server:call({taskrunner_gate, 'runner@localhost'}, {submit, <<"t">>, 10}).
ok
```

`ping` does not test the connection. It sends `{is_auth, ...}` to the
`net_kernel` process on the other node and waits for `yes`. An Ergo node has no
`net_kernel`, so nothing answers and `ping` reports `pang` - then disconnects,
which is why you may see a connection appear and go before the real one. Judge
the connection by whether messages arrive, not by `ping`.

**A cookie mismatch looks like a dead node.** Erlang says:

```
** exception exit: {{nodedown,runner@localhost}, ...}
```

while the Ergo side logs the truth:

```
[warning] BC68B0F0: unable to handshake with 127.0.0.1:56398: malformed DIST handshake ('r' invalid reply)
```

Both nodes must carry the same cookie: `-setcookie taskrunner` on one side,
`options.Network.Cookie` on the other.

## Why there is no observer here

Every other step in the tour keeps the observer running. This one does not.
Observer inspects an Ergo node through the `system` application that every Ergo
node runs - and an Erlang node has none, so there is nothing on the other side
for it to read. The Erlang boundary is a job for the Erlang tools:
`observer:start()`, `recon`, the shell.

## Files

| File | What is in it |
|------|---------------|
| `cmd/runner/main.go` | The three components that make the node speak Erlang |
| `cmd/runner/gate.go` | The one process that speaks ETF, and the type assertions that matter |

Everything under `apps/taskrunner` is exactly as step 08 left it.

## Documentation

- [Erlang](https://docs.ergo.services/extra-library/network-protocols/erlang) -
  EPMD, the handshake, DIST, the full ETF type mapping, and `GenServer`
- [Network Stack](https://docs.ergo.services/networking/network-stack) - what
  the three components replace, and what a network stack is made of
- [Project Structure](https://docs.ergo.services/basics/project-structure) - the
  package API the gate translates into

## Next

[10 - An external program](../10-port) - part of the work done by a program that
is not yours, in a language that is not Go.
