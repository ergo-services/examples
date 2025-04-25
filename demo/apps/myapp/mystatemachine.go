package myapp

import (
	"time"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

// Code Lock, inspired by https://www.erlang.org/doc/system/statem.html#example-revisited

type CodeLockData struct {
	buttons []int
	code    [4]int
}

type CodeLock struct {
	act.StateMachine[CodeLockData]
}

func factoryCodeLock() gen.ProcessBehavior {
	return &CodeLock{}
}

type ButtonPress struct {
	Button int
}

type Lock struct{}

type ResetCode struct{}

type CodeLength struct{}

func (order *CodeLock) Init(args ...any) (act.StateMachineSpec[CodeLockData], error) {
	spec := act.NewStateMachineSpec(
		// The initial state.
		gen.Atom("locked"),

		// The initial data.
		act.WithData(CodeLockData{}),

		// Register a function that is called on every state change.
		act.WithStateEnterCallback(enterState),

		// Register a handler for the ButtonPress message
		act.WithStateMessageHandler(gen.Atom("open"), handleButtonPress),

		// Register the handler for the Lock message.
		act.WithStateMessageHandler(gen.Atom("open"), handleLock),

		// Register the handler for the ResetCode message.
		act.WithStateMessageHandler(gen.Atom("open"), handleResetCode),

		// Register handler for retrieving the length of the code (it would
		// be nice to have an all-state handler for this)
		act.WithStateCallHandler(gen.Atom("open"), handleCodeLength),
		act.WithStateCallHandler(gen.Atom("locked"), handleCodeLength),
	)

	return spec, nil
}

func enterState(oldState gen.Atom, newState gen.Atom, data CodeLockData, proc gen.Process) (gen.Atom, CodeLockData, error) {
	proc.Log().Info("state changed to %s", newState)
	return newState, data, nil
}

func handleButtonPress(state gen.Atom, data CodeLockData, msg ButtonPress, proc gen.Process) (gen.Atom, CodeLockData, []act.Action, error) {
	data.buttons = append(data.buttons, msg.Button)
	if len(data.buttons) < len(data.code) {
		// code incomplete
		// reset code after 30 seconds of inactivity
		reset := act.MessageTimeout{
			Duration: 30 * time.Second,
			Message:  ResetCode{},
		}
		return state, data, []act.Action{reset}, nil
	}
	for i, b := range data.buttons {
		if b != data.code[i] {
			// incorrect code
			// reset buttons
			data.buttons = nil
			return state, data, nil, nil
		}
	}
	// code is correct
	// reset buttons
	// automatically lock after 10 seconds
	data.buttons = nil
	lockAfter10Seconds := act.StateTimeout{
		Duration: 10 * time.Second,
		Message:  Lock{},
	}
	return gen.Atom("open"), data, []act.Action{lockAfter10Seconds}, nil
}

func handleLock(state gen.Atom, data CodeLockData, msg Lock, proc gen.Process) (gen.Atom, CodeLockData, []act.Action, error) {
	doLock()
	data.buttons = nil
	return gen.Atom("locked"), data, nil, nil
}

func handleResetCode(state gen.Atom, data CodeLockData, msg ResetCode, proc gen.Process) (gen.Atom, CodeLockData, []act.Action, error) {
	data.buttons = nil
	return state, data, nil, nil
}

func handleCodeLength(state gen.Atom, data CodeLockData, msg Lock, proc gen.Process) (gen.Atom, CodeLockData, int, []act.Action, error) {
	doLock()
	return state, data, len(data.code), nil, nil
}

func doLock() {
	// interact with the hardware
}

func doUnlock() {
	// interact with the hardware
}
