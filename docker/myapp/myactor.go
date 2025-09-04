package myapp

import (
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/registrar/etcd"
)

// myActor is a simple actor that responds to messages and monitors registrar events
type myActor struct {
	act.Actor
	monitoring bool // Track if we're monitoring events
}

// Init is called when the actor starts
func (a *myActor) Init(args ...any) error {
	a.Log().Info("myActor started on node %s", a.Node().Name())

	// Monitoring/Linking are not allowed here since this process is not fully initialized
	a.Send(a.PID(), "init")

	// Send a periodic heartbeat every 10 seconds
	a.SendAfter(a.PID(), "heartbeat", 10*time.Second)

	return nil
}

// HandleMessage processes incoming messages
func (a *myActor) HandleMessage(from gen.PID, message any) error {
	switch msg := message.(type) {
	case string:
		switch msg {
		case "init":
			// Set up monitoring for registrar events
			if err := a.setupRegistrarMonitoring(); err != nil {
				a.Log().Error("Failed to setup registrar monitoring: %v", err)
			}
			return nil

		case "heartbeat":
			a.Log().Info("myActor heartbeat on node %s", a.Node().Name())
			// Schedule next heartbeat
			a.SendAfter(a.PID(), "heartbeat", 10*time.Second)

		case "ping":
			// Respond to ping with pong
			a.Send(from, "pong")
			a.Log().Info("myActor responded with pong to %s", from)

		case "show_config":
			// Manual configuration display
			a.showCurrentConfig()

		default:
			a.Log().Info("myActor received string message: %s", msg)
		}

	default:
		a.Log().Info("myActor received unknown message type: %T", message)
	}

	return nil
}

// setupRegistrarMonitoring sets up monitoring for registrar events
func (a *myActor) setupRegistrarMonitoring() error {
	registrar, err := a.Node().Network().Registrar()
	if err != nil {
		return err
	}

	// Get the event name for registrar events
	event, err := registrar.Event()
	if err != nil {
		return err
	}

	// Monitor the registrar events using MonitorEvent
	events, err := a.MonitorEvent(event)
	if err != nil {
		return err
	}

	a.monitoring = true
	a.Log().Info("🔍 myActor started monitoring registrar events %s", event.Name)

	// Process any existing events that were returned
	if len(events) > 0 {
		a.Log().Info("📦 Processing %d existing events", len(events))
		for _, existingEvent := range events {
			a.HandleEvent(existingEvent)
		}
	}

	return nil
}

// HandleEvent processes events from monitored processes (registrar)
func (a *myActor) HandleEvent(event gen.MessageEvent) error {
	// The actual event data is in event.Message field
	switch e := event.Message.(type) {
	case etcd.EventConfigUpdate:
		a.handleConfigEvent(e)
	case etcd.EventNodeLeft:
		a.Log().Info("🔻 Node left cluster: %s", e.Name)
	case etcd.EventNodeJoined:
		a.Log().Info("🔺 Node joined cluster: %s", e.Name)
	case etcd.EventApplicationStopped:
		a.Log().Info("🛑 Application stopped: %s on %s", e.Name, e.Node)
	case etcd.EventApplicationStarted:
		a.Log().Info("🚀 Application started: %s on %s", e.Name, e.Node)
	case etcd.EventApplicationLoaded:
		a.Log().Info("📦 Application loaded: %s on %s", e.Name, e.Node)
	case etcd.EventApplicationStopping:
		a.Log().Info("⏹️  Application stopping: %s on %s", e.Name, e.Node)
	case etcd.EventApplicationUnloaded:
		a.Log().Info("📤 Application unloaded: %s on %s", e.Name, e.Node)
	default:
		a.Log().Debug("📨 myActor received event: %T", event.Message)
	}
	return nil
}

// handleConfigEvent processes configuration update events
func (a *myActor) handleConfigEvent(event etcd.EventConfigUpdate) {
	a.Log().Info("⚙️  myActor received configuration update:")
	a.Log().Info("   🔑 Item: %s", event.Item)
	a.Log().Info("   📄 Value: %v (type: %T)", event.Value, event.Value)

	// Handle specific configuration changes
	switch event.Item {
	case "heartbeat.interval":
		if interval, ok := event.Value.(int64); ok {
			a.Log().Info("   💓 Heartbeat interval updated to %d seconds", interval)
			// Note: In production, you might adjust the heartbeat timing
		}
	case "cache.size":
		if size, ok := event.Value.(int64); ok {
			a.Log().Info("   💾 Cache size updated to %d", size)
		}
	case "cache.ratio":
		if ratio, ok := event.Value.(float64); ok {
			a.Log().Info("   📊 Cache ratio updated to %.2f", ratio)
		}
	case "database.host":
		if host, ok := event.Value.(string); ok {
			a.Log().Info("   🏠 Database host updated to %s", host)
		}
	case "database.port":
		if port, ok := event.Value.(int64); ok {
			a.Log().Info("   🔌 Database port updated to %d", port)
		}
	case "log.level":
		if level, ok := event.Value.(string); ok {
			a.Log().Info("   🎚️  Log level updated to %s", level)
			a.Log().Info("   ℹ️  Note: Log level change requires node restart to take effect")
		}
	}

	// Demonstrate real-time configuration access
	a.showCurrentConfig()
}

// showCurrentConfig displays current configuration for this actor
func (a *myActor) showCurrentConfig() {
	registrar, err := a.Node().Network().Registrar()
	if err != nil {
		return
	}

	nodeName := a.Node().Name().String()

	a.Log().Info("   🔄 Current myActor configuration:")

	// Check heartbeat interval
	if heartbeat, err := registrar.ConfigItem("heartbeat.interval"); err == nil {
		a.Log().Info("      💓 heartbeat.interval: %v", heartbeat)
	}

	// Check cache configuration
	if cacheSize, err := registrar.ConfigItem("cache.size"); err == nil {
		a.Log().Info("      💾 cache.size: %v", cacheSize)
	}

	if cacheRatio, err := registrar.ConfigItem("cache.ratio"); err == nil {
		a.Log().Info("      📊 cache.ratio: %v", cacheRatio)
	}

	// Check node-specific config
	config, err := registrar.Config(nodeName+".database.host", nodeName+".database.port")
	if err == nil {
		if dbHost, exists := config[nodeName+".database.host"]; exists {
			a.Log().Info("      🏠 database.host: %v", dbHost)
		}
		if dbPort, exists := config[nodeName+".database.port"]; exists {
			a.Log().Info("      🔌 database.port: %v", dbPort)
		}
	}
}

// HandleCall processes synchronous calls
func (a *myActor) HandleCall(from gen.PID, ref gen.Ref, message any) (any, error) {
	a.Log().Info("myActor received call from %s: %v", from, message)

	if msg, ok := message.(string); ok {
		switch msg {
		case "status":
			return map[string]any{
				"node":       a.Node().Name(),
				"pid":        a.PID(),
				"status":     "running",
				"time":       time.Now().Unix(),
				"monitoring": a.monitoring,
			}, nil

		case "info":
			return map[string]any{
				"actor":      "myActor",
				"node":       a.Node().Name(),
				"app":        "myapp",
				"monitoring": "registrar_events",
			}, nil

		case "config":
			// Return current configuration
			registrar, err := a.Node().Network().Registrar()
			if err != nil {
				return nil, err
			}

			nodeName := a.Node().Name().String()
			configItems := []string{"heartbeat.interval", "cache.size", "cache.ratio",
				nodeName + ".database.host", nodeName + ".database.port"}

			config, err := registrar.Config(configItems...)
			if err != nil {
				return nil, err
			}

			return config, nil

		default:
			return nil, gen.ErrUnsupported
		}
	}

	return nil, gen.ErrUnsupported
}

// Terminate is called when the actor is stopping
func (a *myActor) Terminate(reason error) {
	if a.monitoring {
		a.Log().Info("🔍 Stopping monitoring of registrar events")
		a.monitoring = false
	}

	a.Log().Info("myActor terminating on node %s, reason: %v", a.Node().Name(), reason)
}
