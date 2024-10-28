package main

import (
	"github.com/kohmebot/moyu/moyu"
	"github.com/kohmebot/plugin"
)

func NewPlugin() plugin.Plugin {
	return moyu.NewPluginMoyu()
}
