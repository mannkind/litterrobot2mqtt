package main

type comms struct {
	stateInput    <-chan sourceRep
	stateOutput   chan<- sourceRep
	commandInput  <-chan commandRep
	commandOutput chan<- commandRep
}

func newComms() comms {
	s := make(chan sourceRep, 100)
	c := make(chan commandRep, 100)
	return comms{
		s,
		s,
		c,
		c,
	}
}
