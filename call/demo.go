package main

import (
	"time"

	"ergo.services/ergo"
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/net/edf"
	"ergo.services/logger/colored"
)

type doCallLocal struct{}
type doCallRemote struct{}

type MyMessage struct {
	MyBool   bool
	MyString string
}

func main() {
	var options gen.NodeOptions

	// disable default logger to get rid of multiple logging to the os.Stdout
	options.Log.DefaultLogger.Disable = true

	// add logger "colored".
	loggercolored, err := colored.CreateLogger(colored.Options{})
	if err != nil {
		panic(err)
	}
	options.Log.Loggers = append(options.Log.Loggers, gen.Logger{Name: "cl", Logger: loggercolored})

	options.Log.Level = gen.LogLevelInfo
	// options.Log.Level = gen.LogLevelTrace

	// set network cookie
	options.Network.Cookie = "123"
	// to be able to use self-signed cert
	options.Network.InsecureSkipVerify = true

	// starting node1
	node1, err := ergo.StartNode("node1@localhost", options)
	if err != nil {
		panic(err)
	}

	// register network messages
	if err := edf.RegisterTypeOf(MyMessage{}); err != nil {
		panic(err)
	}

	// use the same options, but remove loggers we added for node1 to use the default one in node2
	options.Log.Loggers = nil
	options.Log.DefaultLogger.Disable = false
	node2, err := ergo.StartNode("node2@localhost", options)
	if err != nil {
		panic(err)
	}
	defer node2.StopForce()

	node2.SpawnRegister("b", factoryB, gen.ProcessOptions{})

	node1.SpawnRegister("b", factoryB, gen.ProcessOptions{})
	node1.SpawnRegister("a", factoryA, gen.ProcessOptions{})

	node1.Send(gen.Atom("a"), doCallLocal{})

	node1.Wait()
}

//
// Actor A
//

func factoryA() gen.ProcessBehavior {
	return &actorA{}
}

type actorA struct {
	act.Actor
}

func (a *actorA) Init(args ...any) error {
	a.Log().Info("started %s process on: %s", a.Name(), a.Node().Name())
	return nil
}

func (a *actorA) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case doCallLocal:
		local := gen.Atom("b")
		if result, err := a.Call(local, MyMessage{MyString: "abc"}); err == nil {
			a.Log().Info("received result from local process %s: %#v", local, result)
		} else {
			a.Log().Error("call local process failed: %s", err)
		}
		a.SendAfter(a.PID(), doCallRemote{}, time.Second)
		return nil

	case doCallRemote:
		remote := gen.ProcessID{Name: "b", Node: "node2@localhost"}
		if result, err := a.Call(remote, MyMessage{MyBool: true, MyString: "def"}); err == nil {
			a.Log().Info("received result from remote process %s: %#v", remote, result)
		} else {
			a.Log().Error("call remote process failed: %s", err)
		}

		a.SendAfter(a.PID(), doCallLocal{}, time.Second)
		return nil

	}

	a.Log().Error("unknown message %#v", message)
	return nil
}

//
// Actor B
//

func factoryB() gen.ProcessBehavior {
	return &actorB{}
}

type actorB struct {
	act.Actor
}

func (b *actorB) Init(args ...any) error {
	b.Log().Info("started %s process on: %s", b.Name(), b.Node().Name())
	return nil
}

func (b *actorB) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	switch r := request.(type) {
	case MyMessage:
		b.Log().Info("received MyMessage request from %s: %#v", from, r)
		if r.MyBool {
			return 1, nil
		}
		return 2, nil
	}

	b.Log().Info("received unknown request: %#v", request)
	return false, nil
}
