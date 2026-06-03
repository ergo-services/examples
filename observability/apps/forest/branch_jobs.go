package forest

import (
	"math/rand"
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

type jobsSupervisor struct {
	act.Supervisor
	id uint64
}

type messageSpawnJob struct{}

func factoryJobsSup() gen.ProcessBehavior {
	return &jobsSupervisor{}
}

func (s *jobsSupervisor) Init(args ...any) (act.SupervisorSpec, error) {
	s.SendAfter(s.PID(), messageSpawnJob{}, time.Second)
	return act.SupervisorSpec{
		Type: act.SupervisorTypeSimpleOneForOne,
		Restart: act.SupervisorRestart{
			Strategy:  act.SupervisorStrategyTemporary,
			Intensity: 100,
			Period:    10,
		},
		Children: []act.SupervisorChildSpec{
			{Name: jobsChild, Factory: factoryJob},
		},
	}, nil
}

func (s *jobsSupervisor) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messageSpawnJob:
		burst := 2 + rand.Intn(6)
		for i := 0; i < burst; i++ {
			s.id++
			if err := s.StartChild(jobsChild, MessageJob{ID: s.id}); err != nil {
				s.Log().Warning("jobs spawn failed: %s", err)
			}
		}
		s.SendAfter(s.PID(), messageSpawnJob{}, time.Duration(700+rand.Intn(1500))*time.Millisecond)
	}
	return nil
}

type jobProc struct {
	act.Actor
	job MessageJob
}

func factoryJob() gen.ProcessBehavior {
	return &jobProc{}
}

func (j *jobProc) Init(args ...any) error {
	if len(args) > 0 {
		if m, ok := args[0].(MessageJob); ok {
			j.job = m
		}
	}
	j.SendAfter(j.PID(), messageJobDone{}, time.Duration(200+rand.Intn(2000))*time.Millisecond)
	return nil
}

type messageJobDone struct{}

func (j *jobProc) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case messageJobDone:
		j.Log().Debug("job %d completed", j.job.ID)
		return gen.TerminateReasonNormal
	}
	return nil
}
