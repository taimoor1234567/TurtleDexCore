package consensus

import (
	"github.com/turtledex/TurtleDexCore/modules"
)

// Alerts implements the Alerter interface for the consensusset.
func (c *ConsensusSet) Alerts() (crit, err, warn []modules.Alert) {
	return []modules.Alert{}, []modules.Alert{}, []modules.Alert{}
}
