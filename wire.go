//+build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/mannkind/twomqtt"
)

func initialize() *app {
	wire.Build(
		newOpts,
		newApp,
		newComms,
		newSink,
		newSource,
		wire.FieldsOf(new(comms), "stateInput"),
		wire.FieldsOf(new(comms), "stateOutput"),
		wire.FieldsOf(new(comms), "commandInput"),
		wire.FieldsOf(new(comms), "commandOutput"),
		wire.FieldsOf(new(opts), "Sink"),
		wire.FieldsOf(new(opts), "Source"),
		wire.FieldsOf(new(sinkOpts), "MQTTOpts"),
		twomqtt.NewMQTT,
	)

	return &app{}
}
