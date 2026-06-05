package forest

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

// deep branch: a deep (~18 levels) + wide recursive supervision tree with long
// structured process names, mirroring observer-example's deeptree. Exercises the
// observer supervision-tree window: deep nesting, large groups that fold into +N
// pills, drill-in into subtrees, and name truncation.

const deepMaxDepth = 18

var deepCounter atomic.Uint64

// generic, domain-neutral pools for long structured names.
var (
	deepRegions  = []string{"us-east", "us-west", "eu-central", "eu-west", "ap-south", "ap-northeast", "sa-east", "me-west"}
	deepServices = []string{"ingest", "transform", "aggregate", "index", "dispatch", "reconcile", "replicate", "compact"}
	deepTiers    = []string{"primary", "replica", "standby", "edge", "batch", "canary"}
)

func deepPick(s []string, n *uint64) string {
	v := s[*n%uint64(len(s))]
	*n /= uint64(len(s))
	return v
}

// deepName builds a unique, long, structured name from the global counter via
// mixed-radix decomposition. region/service/tier vary for visual texture and the
// remaining counter becomes a unique shard id (the distinguishing tail).
// e.g. worker_eu-central:cluster:aggregate:replica_shard-01337
func deepName(role string) gen.Atom {
	n := deepCounter.Add(1)
	region := deepPick(deepRegions, &n)
	service := deepPick(deepServices, &n)
	tier := deepPick(deepTiers, &n)
	return gen.Atom(fmt.Sprintf("%s_%s:cluster:%s:%s_shard-%05d", role, region, service, tier, n))
}

// deepBranch is a supervisor used recursively. Init args: [level, depth, spine].
// `spine` is true only on the main backbone, so the wide fan-out and big leaf
// group apply only there and side subtrees stay small.
type deepBranch struct {
	act.Supervisor
}

func factoryDeepBranch() gen.ProcessBehavior {
	return &deepBranch{}
}

func (s *deepBranch) Init(args ...any) (act.SupervisorSpec, error) {
	level := args[0].(int)
	depth := args[1].(int)
	spine := args[2].(bool)

	children := []act.SupervisorChildSpec{}

	// leaf workers
	leaves := 4
	if spine && level == 6 {
		leaves = 200 // big leaf group -> folds into a +N pill
	}
	for i := 0; i < leaves; i++ {
		children = append(children, act.SupervisorChildSpec{
			Name:    deepName("worker"),
			Factory: factoryDeepLeaf,
		})
	}

	// wide fan-out of supervisors, each with its own shallow subtree -> tests
	// pills of branches and drill-in into one child's subtree.
	if spine && level == 3 {
		for i := 0; i < 200; i++ {
			children = append(children, act.SupervisorChildSpec{
				Name:    deepName("sup"),
				Factory: factoryDeepBranch,
				Args:    []any{level + 1, level + 3, false},
			})
		}
	}

	// deep spine: one supervisor a level deeper
	if level < depth {
		children = append(children, act.SupervisorChildSpec{
			Name:    deepName("sup"),
			Factory: factoryDeepBranch,
			Args:    []any{level + 1, depth, spine},
		})
	}

	return act.SupervisorSpec{
		Type: act.SupervisorTypeOneForOne,
		Restart: act.SupervisorRestart{
			Strategy:  act.SupervisorStrategyPermanent,
			Intensity: 1000,
			Period:    5,
		},
		Children: children,
	}, nil
}

// deepLeaf is a worker that produces a steady, per-worker self-traffic rate so the
// observer shows live activity. Rates are spread across classes so the throughput
// heatmap displays the full bucket range (idle / green / gold / orange / red);
// the same traffic also drives the utilization and mailbox heatmaps.
type deepLeaf struct {
	act.Actor
	rate int // self-messages per second (counts as both in and out)
}

type deepTick struct{}
type deepPing struct{}

const deepTickHz = 5 // ticks per second

func factoryDeepLeaf() gen.ProcessBehavior {
	return &deepLeaf{}
}

func (w *deepLeaf) Init(args ...any) error {
	// weighted classes: mostly idle/low, a few hot — gives a colorful spread
	// without flooding. throughput shown ≈ 2*rate (in + out).
	switch r := rand.Intn(100); {
	case r < 55:
		w.rate = 0 // idle
	case r < 78:
		w.rate = 5 + rand.Intn(45) // ~10-100/s   → green
	case r < 92:
		w.rate = 80 + rand.Intn(320) // ~160-800/s → gold
	case r < 99:
		w.rate = 700 + rand.Intn(1500) // ~1.4-4k/s → orange
	default:
		w.rate = 6000 + rand.Intn(6000) // ~12-24k/s → red (~1%)
	}
	w.SendAfter(w.PID(), deepTick{}, time.Second/deepTickHz)
	return nil
}

func (w *deepLeaf) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case deepTick:
		per := w.rate / deepTickHz // spread the per-second rate across ticks
		for i := 0; i < per; i++ {
			w.Send(w.PID(), deepPing{})
		}
		w.SendAfter(w.PID(), deepTick{}, time.Second/deepTickHz)
	case deepPing:
		// cheap unit of work; just consume
	}
	return nil
}
