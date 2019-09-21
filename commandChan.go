package main

type commandChannel = chan struct {
	Command int64
	State   litterRobotState
}

func newCommandChannel() commandChannel {
	return make(commandChannel, 100)
}
