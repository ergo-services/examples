package taskrunner

import (
	"fmt"
	"os"
	"time"

	"ergo.services/ergo/app"
	"ergo.services/ergo/gen"
)

const resultsFile = "results.log"

// CreateApp returns the application behavior.
func CreateApp() gen.ApplicationBehavior {
	return &taskrunnerApp{}
}

type taskrunnerApp struct {
	app.Application

	results *os.File
}

func (a *taskrunnerApp) Load(args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        "taskrunner",
		Description: "runs tasks and keeps a record of them",
		Mode:        gen.ApplicationModeTransient,

		// The group is what the application watches. One member here: the
		// supervisor. The worker under it is not a member, so the application
		// never hears about a worker restart.
		Group: []gen.ApplicationMemberSpec{
			{Name: supName, Factory: factorySupervisor},
		},

		// The node that hosts the application learns the wire types from here,
		// when the application is loaded. Nothing to remember in main.
		Network: gen.ApplicationNetwork{
			RegisterTypes: wireTypes,
		},
	}, nil
}

// Init runs before a single member exists. Whatever the group needs to work
// has to be ready when it returns.
func (a *taskrunnerApp) Init(ref gen.Ref, mode gen.ApplicationMode) error {
	f, err := os.OpenFile(resultsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	a.results = f

	a.Log().Info("init: opened %s, no member is running yet", resultsFile)
	return nil
}

// Start runs once every member is spawned.
func (a *taskrunnerApp) Start(ref gen.Ref, mode gen.ApplicationMode) {
	a.Log().Info("start: the group is up, mode is %s", mode)
}

// Stop runs before the surviving members are asked to exit. It is the place to
// stop accepting work, not to release anything the group still uses.
func (a *taskrunnerApp) Stop(ref gen.Ref, reason error) {
	a.Log().Info("stop: going down (%s)", reason)
}

// Terminate runs after the last member is gone. Nothing can use the file any
// more, so this is where it is closed.
func (a *taskrunnerApp) Terminate(reason error) {
	a.Log().Info("terminate: the group is gone, closing %s", resultsFile)
	a.results.Close()
}

// Record writes one line to the results file. The worker calls it through
// process.Application().Behavior(), which is how a group process reaches
// something the application owns.
//
// One worker means one writer. Step 06 adds a pool, and there the writes go
// through a single actor instead.
func (a *taskrunnerApp) Record(id string, took time.Duration) {
	fmt.Fprintf(a.results, "%s %s %s\n", time.Now().Format(time.TimeOnly), id, took)
}
