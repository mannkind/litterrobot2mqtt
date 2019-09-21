package main

type stateChannel = chan litterRobotState

func newStateChannel() stateChannel {
	return make(stateChannel, 100)
}
