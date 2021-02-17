package build

var (
	// siaAPIPassword is the environment variable that sets a custom API
	// password if the default is not used
	siaAPIPassword = "SIA_API_PASSWORD"

	// siaDataDir is the environment variable that tells ttdxd where to put the
	// general sia data, e.g. api password, configuration, logs, etc.
	siaDataDir = "SIA_DATA_DIR"

	// ttdxdDataDir is the environment variable which tells ttdxd where to put the
	// ttdxd-specific data
	ttdxdDataDir = "SIAD_DATA_DIR"

	// siaWalletPassword is the environment variable that can be set to enable
	// auto unlocking the wallet
	siaWalletPassword = "SIA_WALLET_PASSWORD"

	// siaExchangeRate is the environment variable that can be set to
	// show amounts (additionally) in a different currency
	siaExchangeRate = "SIA_EXCHANGE_RATE"
)
