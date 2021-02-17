package main

import (
	"strings"

	"github.com/turtledex/TurtleDexCore/node"
)

// createNodeParams parses the provided config and creates the corresponding
// node params for the server.
func parseModules(config Config) node.NodeParams {
	params := node.NodeParams{}
	// Parse the modules.
	if strings.Contains(config.TurtleDexd.Modules, "g") {
		params.CreateGateway = true
	}
	if strings.Contains(config.TurtleDexd.Modules, "c") {
		params.CreateConsensusSet = true
	}
	if strings.Contains(config.TurtleDexd.Modules, "e") {
		params.CreateExplorer = true
	}
	if strings.Contains(config.TurtleDexd.Modules, "f") {
		params.CreateFeeManager = true
	}
	if strings.Contains(config.TurtleDexd.Modules, "t") {
		params.CreateTransactionPool = true
	}
	if strings.Contains(config.TurtleDexd.Modules, "w") {
		params.CreateWallet = true
	}
	if strings.Contains(config.TurtleDexd.Modules, "m") {
		params.CreateMiner = true
	}
	if strings.Contains(config.TurtleDexd.Modules, "h") {
		params.CreateHost = true
	}
	if strings.Contains(config.TurtleDexd.Modules, "r") {
		params.CreateRenter = true
	}
	// Parse remaining fields.
	params.Bootstrap = !config.TurtleDexd.NoBootstrap
	params.HostAddress = config.TurtleDexd.HostAddr
	params.RPCAddress = config.TurtleDexd.RPCaddr
	params.TurtleDexMuxTCPAddress = config.TurtleDexd.TurtleDexMuxTCPAddr
	params.TurtleDexMuxWSAddress = config.TurtleDexd.TurtleDexMuxWSAddr
	params.Dir = config.TurtleDexd.TurtleDexDir
	return params
}
