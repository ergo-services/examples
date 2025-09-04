package main

import (
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/registrar/etcd"
)

// clientActor resolves myapp and connects to the first available route
type clientActor struct {
	act.Actor
	targetNode gen.Atom
	connected  bool
	monitoring bool // Track if we're monitoring events
}

func (c *clientActor) Init(args ...any) error {
	c.Log().Info("clientActor started on node %s", c.Node().Name())

	// Monitoring/Linking are not allowed here since this process is not fully initialized
	c.Send(c.PID(), "init")

	// Start the resolution process after a short delay to let other nodes start
	c.SendAfter(c.PID(), "resolve_app", 5*time.Second)

	return nil
}

func (c *clientActor) HandleMessage(from gen.PID, message any) error {
	switch msg := message.(type) {
	case string:
		switch msg {
		case "init":
			// Set up monitoring for registrar events
			if err := c.setupRegistrarMonitoring(); err != nil {
				c.Log().Error("Failed to setup registrar monitoring: %v", err)
			}
			return nil

		case "resolve_app":
			c.resolveAndConnect()

		case "ping_target":
			c.pingTarget()

		case "pong":
			c.Log().Info("clientActor received pong response!")
			// Schedule next ping in 15 seconds
			c.SendAfter(c.PID(), "ping_target", 15*time.Second)

		case "show_config":
			// Manual configuration display
			c.showCurrentConfig()

		case "reconnect":
			// Manual reconnection trigger
			c.connected = false
			c.targetNode = gen.Atom("")
			c.SendAfter(c.PID(), "resolve_app", 1*time.Second)

		default:
			c.Log().Info("clientActor received string message: %s", msg)
		}

	default:
		c.Log().Info("clientActor received unknown message type: %T", message)
	}

	return nil
}

// setupRegistrarMonitoring sets up monitoring for registrar events
func (c *clientActor) setupRegistrarMonitoring() error {
	registrar, err := c.Node().Network().Registrar()
	if err != nil {
		return err
	}

	// Get the event name for registrar events
	event, err := registrar.Event()
	if err != nil {
		return err
	}

	// Monitor the registrar events using MonitorEvent
	events, err := c.MonitorEvent(event)
	if err != nil {
		return err
	}

	c.monitoring = true
	c.Log().Info("🔍 clientActor started monitoring registrar events %s", event.Name)

	// Process any existing events that were returned
	if len(events) > 0 {
		c.Log().Info("📦 Processing %d existing events", len(events))
		for _, existingEvent := range events {
			c.HandleEvent(existingEvent)
		}
	}

	return nil
}

// HandleEvent processes events from monitored processes (registrar)
func (c *clientActor) HandleEvent(event gen.MessageEvent) error {
	// The actual event data is in event.Message field
	switch e := event.Message.(type) {
	case etcd.EventConfigUpdate:
		c.handleConfigEvent(e)
	case etcd.EventNodeLeft:
		c.Log().Info("🔻 Node left cluster: %s", e.Name)
		// Check if our target node went down
		if c.targetNode == e.Name {
			c.Log().Warning("📍 Target node %s went down! Clearing connection", e.Name)
			c.connected = false
			c.targetNode = gen.Atom("")
			// Try to reconnect after a delay
			c.SendAfter(c.PID(), "resolve_app", 3*time.Second)
		}
	case etcd.EventNodeJoined:
		c.Log().Info("🔺 Node joined cluster: %s", e.Name)
		// If we're not connected, try to resolve applications
		if !c.connected {
			c.Log().Info("🔄 New node joined, attempting to resolve applications")
			c.SendAfter(c.PID(), "resolve_app", 2*time.Second)
		}
	case etcd.EventApplicationStopped:
		c.Log().Info("🛑 Application stopped: %s on %s", e.Name, e.Node)
		// If myapp stopped on our target node, clear connection
		if e.Name == "myapp" && c.targetNode == e.Node {
			c.Log().Warning("📍 myapp stopped on target node %s! Clearing connection", e.Node)
			c.connected = false
			c.targetNode = gen.Atom("")
			c.SendAfter(c.PID(), "resolve_app", 3*time.Second)
		}
	case etcd.EventApplicationStarted:
		c.Log().Info("🚀 Application started: %s on %s", e.Name, e.Node)
		// If myapp started and we're not connected, try to connect
		if e.Name == "myapp" && !c.connected {
			c.Log().Info("🔄 myapp started, attempting to connect")
			c.SendAfter(c.PID(), "resolve_app", 1*time.Second)
		}
	case etcd.EventApplicationLoaded:
		c.Log().Info("📦 Application loaded: %s on %s", e.Name, e.Node)
	case etcd.EventApplicationStopping:
		c.Log().Info("⏹️  Application stopping: %s on %s", e.Name, e.Node)
	case etcd.EventApplicationUnloaded:
		c.Log().Info("📤 Application unloaded: %s on %s", e.Name, e.Node)
	default:
		c.Log().Debug("📨 clientActor received event: %T", event.Message)
	}
	return nil
}

// handleConfigEvent processes configuration update events
func (c *clientActor) handleConfigEvent(event etcd.EventConfigUpdate) {
	c.Log().Info("⚙️  clientActor received configuration update:")
	c.Log().Info("   🔑 Item: %s", event.Item)
	c.Log().Info("   📄 Value: %v (type: %T)", event.Value, event.Value)

	// Handle specific configuration changes
	switch event.Item {
	case "ping.interval":
		if interval, ok := event.Value.(int64); ok {
			c.Log().Info("   🏓 Ping interval updated to %d seconds", interval)
		}
	case "resolver.timeout":
		if timeout, ok := event.Value.(int64); ok {
			c.Log().Info("   ⏱️  Resolver timeout updated to %d seconds", timeout)
		}
	case "resolver.max_retries":
		if retries, ok := event.Value.(int64); ok {
			c.Log().Info("   🔄 Resolver max retries updated to %d", retries)
		}
	case "cache.size":
		if size, ok := event.Value.(int64); ok {
			c.Log().Info("   💾 Cache size updated to %d", size)
		}
	case "log.level":
		if level, ok := event.Value.(string); ok {
			c.Log().Info("   🎚️  Log level updated to %s", level)
		}
	}

	// Show current configuration
	c.showCurrentConfig()
}

// showCurrentConfig displays current configuration for this actor
func (c *clientActor) showCurrentConfig() {
	registrar, err := c.Node().Network().Registrar()
	if err != nil {
		return
	}

	nodeName := c.Node().Name().String()

	c.Log().Info("   🔄 Current clientActor configuration:")

	// Check client-specific config
	if pingInterval, err := registrar.ConfigItem("ping.interval"); err == nil {
		c.Log().Info("      🏓 ping.interval: %v", pingInterval)
	}

	if cacheSize, err := registrar.ConfigItem("cache.size"); err == nil {
		c.Log().Info("      💾 cache.size: %v", cacheSize)
	}

	// Check node-specific config
	config, err := registrar.Config(nodeName+".resolver.timeout", nodeName+".resolver.max_retries")
	if err == nil {
		if resolverTimeout, exists := config[nodeName+".resolver.timeout"]; exists {
			c.Log().Info("      ⏱️  resolver.timeout: %v", resolverTimeout)
		}
		if maxRetries, exists := config[nodeName+".resolver.max_retries"]; exists {
			c.Log().Info("      🔄 resolver.max_retries: %v", maxRetries)
		}
	}
}

// resolveAndConnect resolves the myapp application and connects to the first route
func (c *clientActor) resolveAndConnect() {
	c.Log().Info("Resolving myapp application...")

	// Get the registrar
	registrar, err := c.Node().Network().Registrar()
	if err != nil {
		c.Log().Error("Failed to get registrar: %v", err)
		// Retry after 5 seconds
		c.SendAfter(c.PID(), "resolve_app", 5*time.Second)
		return
	}

	// Get the resolver interface
	resolver := registrar.Resolver()

	// Resolve the myapp application
	routes, err := resolver.ResolveApplication("myapp")
	if err != nil {
		c.Log().Error("Failed to resolve myapp: %v", err)
		// Retry after 5 seconds
		c.SendAfter(c.PID(), "resolve_app", 5*time.Second)
		return
	}

	if len(routes) == 0 {
		c.Log().Warning("No routes found for myapp, retrying...")
		// Retry after 5 seconds
		c.SendAfter(c.PID(), "resolve_app", 5*time.Second)
		return
	}

	// Use the first route
	firstRoute := routes[0]
	c.Log().Info("Found %d routes for myapp, using first route on node %s", len(routes), firstRoute.Node)

	// Store the target node
	c.targetNode = firstRoute.Node
	c.connected = true

	c.Log().Info("Connected to myactor on node %s", c.targetNode)

	// Start pinging the target after connection
	c.SendAfter(c.PID(), "ping_target", 2*time.Second)
}

// pingTarget sends a ping message to the target actor
func (c *clientActor) pingTarget() {
	if !c.connected {
		c.Log().Warning("Not connected, trying to resolve again...")
		c.SendAfter(c.PID(), "resolve_app", 2*time.Second)
		return
	}

	c.Log().Info("Attempting to connect to myactor on node %s", c.targetNode)

	// For now, log that we found the route but can't send directly
	// In a real application, you would use the route information to establish connections
	c.Log().Info("Found myapp route on node %s - in production this would establish a connection", c.targetNode)

	// Schedule next attempt
	c.SendAfter(c.PID(), "ping_target", 15*time.Second)
}

func (c *clientActor) Terminate(reason error) {
	if c.monitoring {
		c.Log().Info("🔍 Stopping monitoring of registrar events")
		c.monitoring = false
	}

	c.Log().Info("clientActor terminating on node %s, reason: %v", c.Node().Name(), reason)
}
