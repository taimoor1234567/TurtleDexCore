package host

import (
	"bytes"
	"path/filepath"

	"github.com/turtledex/TurtleDexCore/build"
	"github.com/turtledex/TurtleDexCore/modules"
	"github.com/turtledex/TurtleDexCore/persist"
	"github.com/turtledex/errors"
)

// upgradeFromV120ToV143 is an upgrade layer that aids the integration of the
// TurtleDexMux. Seeing as the TurtleDexMux should use the host's public and private keys,
// we need a version bump to trigger the TurtleDexMux's compatibility flow. If a node
// starts up and we notice the host's persistence is outdated and needs an
// upgrade, we initialize the TurtleDexMux with the host's key pair.
func (h *Host) upgradeFromV120ToV143() error {
	h.log.Println("Attempting an upgrade for the host from v1.2.0 to v1.4.3")

	// Load the persistence object
	p := new(persistence)
	err := h.dependencies.LoadFile(modules.Hostv120PersistMetadata, p, filepath.Join(h.persistDir, settingsFile))
	if err != nil {
		return build.ExtendErr("could not load persistence object", err)
	}

	// Add the ephemeral account defaults
	p.Settings.EphemeralAccountExpiry = modules.DefaultEphemeralAccountExpiry
	p.Settings.MaxEphemeralAccountBalance = modules.DefaultMaxEphemeralAccountBalance
	p.Settings.MaxEphemeralAccountRisk = defaultMaxEphemeralAccountRisk

	// Load it on the host
	h.loadPersistObject(p)

	// Verify the host and siamux share the same keypair
	smsk := h.staticMux.PrivateKey()
	smpk := h.staticMux.PublicKey()
	if !bytes.Equal(h.secretKey[:], smsk[:]) {
		return errors.New("expected host private key to equal the siamux's private key")
	}
	if !bytes.Equal(h.publicKey.Key[:], smpk[:]) {
		return errors.New("expected host public key to equal the siamux's public key")
	}

	// Save the updated persist so that the upgrade is not triggered again.
	err = persist.SaveJSON(modules.Hostv143PersistMetadata, h.persistData(), filepath.Join(h.persistDir, settingsFile))
	if err != nil {
		return build.ExtendErr("could not save persistence object", err)
	}

	return nil
}
