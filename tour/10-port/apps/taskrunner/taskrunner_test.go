package taskrunner

import (
	"testing"
	"time"

	"ergo.services/ergo/gen"
	"ergo.services/ergo/testing/check"
	"ergo.services/ergo/testing/unit"
)

// fakeRecorder is the whole reason the auditor takes a recorder instead of
// reaching into the application.
type fakeRecorder struct {
	ids []string
}

func (f *fakeRecorder) Record(id string, took time.Duration) {
	f.ids = append(f.ids, id)
}

// The worker does the work and reports it. Nothing about a node, a supervisor
// or an application is involved: unit.Spawn runs the behavior on its own.
func TestWorkerReportsToTheAuditor(t *testing.T) {
	w, err := unit.Spawn(t, factoryWorker, gen.ProcessOptions{})
	check.NoError(t, err)

	from := gen.PID{Node: "client@localhost", ID: 1000}
	w.SendMessage(from, messageSubmit{ID: "task-1", Effort: time.Millisecond})

	// What the worker did is what it sent. The Took field is not asserted: it
	// is a measurement, and a test that pins it down would fail on a busy
	// machine for no reason.
	rec, ok := w.ShouldSend().To(auditorName).Capture()
	check.True(t, ok)

	done, ok := rec.Message.(messageDone)
	check.True(t, ok)
	check.Equal(t, "task-1", done.ID)
	check.Equal(t, "client@localhost", done.For)

	w.ShouldTerminate().None().Assert()
}

// The auditor registers the event with the options it needs. Getting Buffer
// wrong is invisible until a subscriber arrives late and finds nothing.
func TestAuditorRegistersWithNotifyAndBuffer(t *testing.T) {
	a, err := unit.Spawn(t, factoryAuditor, gen.ProcessOptions{}, &fakeRecorder{})
	check.NoError(t, err)

	a.ShouldRegisterEvent().Name(eventName).Notify(true).Buffer(10).Once().Assert()
}

// One report in, one line recorded and one event out.
func TestAuditorRecordsAndPublishes(t *testing.T) {
	records := &fakeRecorder{}
	a, err := unit.Spawn(t, factoryAuditor, gen.ProcessOptions{}, records)
	check.NoError(t, err)

	a.SendMessage(gen.PID{}, messageDone{
		ID:   "task-1",
		For:  "client@localhost",
		Took: 150 * time.Millisecond,
	})

	check.Equal(t, []string{"task-1"}, records.ids)
	a.ShouldSendEvent().Name(eventName).Message(EventTaskDone{
		ID:   "task-1",
		For:  "client@localhost",
		Took: 150 * time.Millisecond,
	}).Once().Assert()
}

// An auditor spawned without a recorder refuses to start rather than panicking
// on the first report.
func TestAuditorNeedsARecorder(t *testing.T) {
	_, err := unit.Spawn(t, factoryAuditor, gen.ProcessOptions{})
	check.ErrorIs(t, err, gen.ErrIncorrect)
}

// The pool answers a request that arrives above normal priority.
func TestPoolGrowsOnHighPriorityRequest(t *testing.T) {
	p, err := unit.Spawn(t, factoryPool, gen.ProcessOptions{})
	check.NoError(t, err)

	result, err := p.CallWithPriority(gen.PID{}, requestAddWorkers{N: 2}, gen.MessagePriorityHigh)
	check.NoError(t, err)
	check.Equal(t, responseAddWorkers{Size: 5}, result)
}

// And forwards one that arrives at normal priority, which is the trap: the
// caller waits for an answer from a worker that has no idea what the request
// means. This is the test that pins the rule down.
func TestPoolForwardsANormalPriorityRequest(t *testing.T) {
	p, err := unit.Spawn(t, factoryPool, gen.ProcessOptions{})
	check.NoError(t, err)

	result, err := p.Call(gen.PID{}, requestAddWorkers{N: 2})
	check.NoError(t, err)
	check.Nil(t, result)

	p.ShouldForward().Once().Assert()
}
