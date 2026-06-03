package forest

import (
	"ergo.services/ergo/app"
	"ergo.services/ergo/gen"
)

const (
	appName    gen.Atom = "forest_scenario"
	rootSup    gen.Atom = "forest_root"
	computeSup gen.Atom = "forest_compute"
	ingestSup  gen.Atom = "forest_ingest"
	jobsSup    gen.Atom = "forest_jobs"

	computePool        gen.Atom = "forest_compute_pool"
	computeCoordinator gen.Atom = "forest_compute_coordinator"

	ingestRouter     gen.Atom = "forest_ingest_router"
	ingestAggregator gen.Atom = "forest_ingest_aggregator"

	jobsChild gen.Atom = "forest_job"
)

func CreateApp() gen.ApplicationBehavior {
	return &forestApp{}
}

type forestApp struct {
	app.Application
}

func (a *forestApp) Load(args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        appName,
		Description: "Forest scenario: nested supervisors with pool, router, and dynamic jobs",
		Mode:        gen.ApplicationModeTemporary,
		Network: gen.ApplicationNetwork{
			RegisterTypes: []any{
				MessageCompute{},
				MessageIngest{},
				MessageJob{},
			},
		},
		Group: []gen.ApplicationMemberSpec{
			{
				Name:    rootSup,
				Factory: factoryRoot,
			},
		},
	}, nil
}
