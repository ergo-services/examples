package events

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

const watchSupName gen.Atom = "events_watch_sup"

// watchSup statically supervises the singleton watch-window demo events (one-for-one,
// auto-started — unlike the load publishers/subscribers which are dynamic SOFO children).
func factoryWatchSup() gen.ProcessBehavior { return &watchSup{} }

type watchSup struct {
	act.Supervisor
}

func (s *watchSup) Init(args ...any) (act.SupervisorSpec, error) {
	return act.SupervisorSpec{
		Type: act.SupervisorTypeOneForOne,
		Restart: act.SupervisorRestart{
			Strategy:  act.SupervisorStrategyPermanent,
			Intensity: 10,
			Period:    5,
		},
		Children: []act.SupervisorChildSpec{
			{Name: "temperature_sensor_pub", Factory: factoryTemperatureSensor},
			{Name: "chat_room_pub", Factory: factoryChatRoom},
			{Name: "chat_member", Factory: factoryChatMember},
			{Name: "orders_created_pub", Factory: factoryOrdersCreated},
			{Name: "service_heartbeat_pub", Factory: factoryServiceHeartbeat},
			{Name: "audit_trail_pub", Factory: factoryAuditTrail},
			{Name: "deploy_status_pub", Factory: factoryDeployStatus},
		},
	}, nil
}
