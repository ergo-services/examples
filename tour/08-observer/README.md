# 08 - Looking inside

## What this shows

The observer has been running since step 01, in every one of these examples, and
this step is about actually using it. Two ways in:

- **The browser**, at http://localhost:9911. Processes, applications,
  connections, events, the log as it happens.
- **An AI agent**, over MCP on the same port. Thirty-eight tools, of which nine
  change something on the node - and on this node the agent is given only the
  twenty-nine that do not.

The client now keeps the node permanently busy: one task a second, and a burst
of sixty every thirty seconds. There is something to look at whenever you look.

## Run it

Two terminals, in either order:

```bash
go run ./cmd/runner
```

```bash
go run ./cmd/client
```

Then leave them running.

## Find the missing tasks

The client counts what it submitted and what came back, and reports the tally at
the start of each burst, when the previous round has settled:

```
[info] <...1050>: so far: 30 submitted, 30 came back
[info] <...1050>: --- a burst of 60 ---
[info] <...1050>: so far: 120 submitted, 92 came back
```

Thirty out of thirty, then ninety-two out of a hundred and twenty. Twenty-eight
tasks went somewhere. You know from step 06 where - the point here is to find it
without knowing.

Open http://localhost:9911, look at the processes on `runner@localhost`, and
open `taskrunner_worker`. Its state is whatever the process answers about
itself, and a pool answers with its own accounting:

```
ergo:pool_size            3
ergo:worker_mailbox_size  10
ergo:messages_forwarded   92
ergo:messages_unhandled   28
```

That is the whole diagnosis, in the place where a dropped message is the only
thing recorded. Nothing in the worker is broken; there was nowhere to put the
work.

While you are there, the rest of the node is worth a minute. The application
tree shows `taskrunner` with the supervisor, the pool and the auditor under it.
The connections lens shows `client@localhost`. The events lens shows `tasks`
with its subscriber on the other node.

## Point an agent at it

```bash
claude mcp add --transport http ergo-tour http://localhost:9911/mcp
```

Then ask questions in ordinary words. Three that work on this node:

- *"What is running on the runner node, and which of it belongs to the
  taskrunner application?"*
- *"Tasks submitted to runner@localhost are not all coming back. Where do they
  go?"*
- *"Who is subscribed to the tasks event, and from where?"*

The agent has the same data the browser does, addressed as resources rather
than pages: `ergo://cluster` names the nodes, and each node has thirteen lenses
addressed as `ergo://<node>/<lens>[/<target>]` - processes, process, network,
connections, applications, events, log, tracing and the rest. The tools are the
active side of the same surface: list processes, read one's state, follow an
event, take a goroutine dump or a heap profile.

## Read, not touch

`cmd/runner/main.go` spells the listener out instead of taking the port default,
for one reason:

```go
Listeners: []observer.Listener{
	{
		Port: 9911,
		MCP: observer.SurfaceMCP{
			Ceiling:      &observer.Ceiling{ReadOnly: true},
			Instructions: instructions,
		},
	},
},
```

A ceiling on a surface narrows that surface alone: the browser on this listener
keeps everything, and the MCP surface becomes read-only. And read-only does not
mean "asks and is refused" - the nine mutating tools are not in the list the
agent is given at all, so `kill` and `app_stop` are not options it has to be
trusted with.

A surface ceiling can only narrow, never widen. Putting `ReadOnly` on the
listener and a full ceiling on the surface leaves the surface read-only.

## What the agent is told first

`Instructions` is a sentence or two about what this node is, handed to the agent
before it asks anything:

> This node runs the taskrunner application from the Ergo tour. Tasks arrive at
> the registered process taskrunner_worker, which is a pool of workers;
> taskrunner_auditor records every finished task and publishes the 'tasks'
> event. Tasks come from client@localhost.

It is added to the guidance the observer gives about itself, not put in its
place, so the agent still knows how to navigate the surface. Without it an agent
has to discover the shape of the deployment by poking at it; with it, the first
question already lands in the right place.

## One thing to know before deploying this

This listener has no `Authorizer`, so it is open: whoever can reach the port
gets its ceiling. On localhost that is fine and it is why the tour works with no
setup. Anywhere else, a listener needs an authorizer, and a ceiling is the
backstop rather than the whole story - the observer refuses at start if you ask
for something inconsistent.

## Files

| File | What is in it |
|------|---------------|
| `cmd/runner/main.go` | The listener, the read-only MCP surface, the instructions |
| `cmd/client/client.go` | Steady traffic and a burst every thirty seconds |

Everything under `apps/taskrunner` is exactly as step 07 left it.

## Documentation

- [Inspecting With Observer](https://docs.ergo.services/advanced/observer) - the
  lenses, and what each one answers
- [Inspecting With an AI Agent](https://docs.ergo.services/advanced/mcp) - the
  tool catalogue, the `ergo://` resources, and what to ask
- [Observer](https://docs.ergo.services/extra-library/applications/observer) -
  every option: listeners, surfaces, ceilings, authorizers, rate limits
- [Inspecting Actor State](https://docs.ergo.services/advanced/inspecting-state) -
  `HandleInspect`, so your own actors answer as well as the pool does
- [Debugging](https://docs.ergo.services/advanced/debugging) - build tags,
  goroutine dumps and heap profiles, on a node that is misbehaving
- [AI Agents](https://docs.ergo.services/ai-agents) - the skills and agents that
  come with the framework

## Next

[09 - Talking to Erlang](../09-erlang) - the same task runner, driven from an
`erl` shell.
