package gateway

import "github.com/turtledex/TurtleDexCore/modules"

// Alerts implements the modules.Alerter interface for the gateway.
func (g *Gateway) Alerts() (crit, err, warn []modules.Alert) {
	return g.staticAlerter.Alerts()
}
