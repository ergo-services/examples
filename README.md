[![Gitbook Documentation](https://img.shields.io/badge/GitBook-Documentation-f37f40?style=plastic&logo=gitbook&logoColor=white&style=flat)](https://docs.ergo.services)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)
[![Telegram Community](https://img.shields.io/badge/Telegram-ergo__services-229ed9?style=flat&logo=telegram&logoColor=white)](https://t.me/ergo_services)
[![Reddit](https://img.shields.io/badge/Reddit-r/ergo__services-ff4500?style=plastic&logo=reddit&logoColor=white&style=flat)](https://reddit.com/r/ergo_services)

# Ergo Framework examples

## The tour

Ten steps, in order. Each one is its own Go module,
each adds exactly one idea to the one before it, and each `README.md` says what
to look for in the output. The domain is the same throughout: a task runner.

| Step | What it adds |
|------|--------------|
| [01 - The first actor](tour/01-first-actor) | An actor, a mailbox, `Send` and `Call` |
| [02 - When a worker dies](tour/02-supervision) | A supervisor, restarts, and the limit on them |
| [03 - Making it an application](tour/03-application) | An application, its group, and who owns a resource |
| [04 - Two nodes](tour/04-network) | The network: node names, the cookie, wire types, monitors |
| [05 - Events](tour/05-events) | One publisher, several subscribers, and a buffer |
| [06 - A pool of workers](tour/06-pool) | `act.Pool`, and what it does when it cannot keep up |
| [07 - Testing it](tour/07-testing) | `testing/unit` for one actor, `testing/stage` for two nodes |
| [08 - Looking inside](tour/08-observer) | The observer, and an AI agent over MCP |
| [09 - Talking to Erlang](tour/09-erlang) | The same code, driven from an `erl` shell |
| [10 - An external program](tour/10-port) | Running a program in another language and talking to it |

Start at 01. Steps 01 to 08 keep the observer on http://localhost:9911, so there
is always a running system to look at, and step 08 is about reading it. Step 09
is the only one that needs anything installed: Erlang.

## observability

Not a lesson but a stand: a five-node cluster under continuous load, with
Prometheus metrics, Grafana dashboards, distributed tracing and the observer, so
that every observability tool has something to work with. Runs with Docker
Compose. See [observability/README.md](observability) for the details.

## Reference

The framework documentation is at https://docs.ergo.services. The examples here
deliberately show one thing at a time; the documentation is where the full API
lives.
