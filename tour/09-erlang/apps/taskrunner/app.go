package taskrunner

import (
	"fmt"
	"os"
	"time"

	"ergo.services/ergo/app"
	"ergo.services/ergo/gen"
)

// Options configures the application. The results path is a field now, not a
// constant: a test has to be able to point it somewhere disposable.
type Options struct {
	Results string
}

// CreateApp returns the application behavior.
func CreateApp(options Options) gen.ApplicationBehavior {
	if options.Results == "" {
		options.Results = "results.log"
	}
	return &taskrunnerApp{options: options}
}

type taskrunnerApp struct {
	app.Application

	options Options
	results *os.File
}

func (a *taskrunnerApp) Load(args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        "taskrunner",
		Description: "runs tasks and keeps a record of them",
		Mode:        gen.ApplicationModeTransient,

		// The group is what the application watches. One member here: the
		// supervisor. Everything under it belongs to the application too, but
		// is not a member.
		//
		// Args carries the application itself down to whoever needs to record
		// something. Handing it over is what makes that process testable: the
		// auditor takes a recorder, and a test can pass its own.
		Group: []gen.ApplicationMemberSpec{
			{Name: supName, Factory: factorySupervisor, Args: []any{a}},
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
	f, err := os.OpenFile(a.options.Results, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	a.results = f

	a.Log().Info("init: opened %s, no member is running yet", a.options.Results)
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
	a.Log().Info("terminate: the group is gone, closing %s", a.options.Results)
	a.results.Close()
}

// Record writes one line to the results file, and satisfies the recorder the
// auditor asks for.
func (a *taskrunnerApp) Record(id string, took time.Duration) {
	fmt.Fprintf(a.results, "%s %s %s\n", time.Now().Format(time.TimeOnly), id, took)
}
