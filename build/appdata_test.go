package build

import (
	"os"
	"testing"
)

// TestAPIPassword tests getting and setting the API Password
func TestAPIPassword(t *testing.T) {
	// Unset any defaults, this only affects in memory state. Any Env Vars will
	// remain intact on disk
	err := os.Unsetenv(siaAPIPassword)
	if err != nil {
		t.Error(err)
	}

	// Calling APIPassword should return a non-blank password if the env
	// variable isn't set
	pw, err := APIPassword()
	if err != nil {
		t.Error(err)
	}
	if pw == "" {
		t.Error("Password should not be blank")
	}

	// Test setting the env variable
	newPW := "abc123"
	err = os.Setenv(siaAPIPassword, newPW)
	if err != nil {
		t.Error(err)
	}
	pw, err = APIPassword()
	if err != nil {
		t.Error(err)
	}
	if pw != newPW {
		t.Errorf("Expected password to be %v but was %v", newPW, pw)
	}
}

// TestTurtleDexdDataDir tests getting and setting the TurtleDex consensus directory
func TestTurtleDexdDataDir(t *testing.T) {
	// Unset any defaults, this only affects in memory state. Any Env Vars will
	// remain intact on disk
	err := os.Unsetenv(ttdxdDataDir)
	if err != nil {
		t.Error(err)
	}

	// Test Default TurtleDexdDataDir
	ttdxdDir := TurtleDexdDataDir()
	if ttdxdDir != "" {
		t.Errorf("Expected ttdxdDir to be empty but was %v", ttdxdDir)
	}

	// Test Env Variable
	newTurtleDexDir := "foo/bar"
	err = os.Setenv(ttdxdDataDir, newTurtleDexDir)
	if err != nil {
		t.Error(err)
	}
	ttdxdDir = TurtleDexdDataDir()
	if ttdxdDir != newTurtleDexDir {
		t.Errorf("Expected ttdxdDir to be %v but was %v", newTurtleDexDir, ttdxdDir)
	}
}

// TestTurtleDexDir tests getting and setting the TurtleDex data directory
func TestTurtleDexDir(t *testing.T) {
	// Unset any defaults, this only affects in memory state. Any Env Vars will
	// remain intact on disk
	err := os.Unsetenv(siaDataDir)
	if err != nil {
		t.Error(err)
	}

	// Test Default TurtleDexDir
	siaDir := TurtleDexDir()
	if siaDir != defaultTurtleDexDir() {
		t.Errorf("Expected siaDir to be %v but was %v", defaultTurtleDexDir(), siaDir)
	}

	// Test Env Variable
	newTurtleDexDir := "foo/bar"
	err = os.Setenv(siaDataDir, newTurtleDexDir)
	if err != nil {
		t.Error(err)
	}
	siaDir = TurtleDexDir()
	if siaDir != newTurtleDexDir {
		t.Errorf("Expected siaDir to be %v but was %v", newTurtleDexDir, siaDir)
	}
}

// TestTurtleDexWalletPassword tests getting and setting the TurtleDex Wallet Password
func TestTurtleDexWalletPassword(t *testing.T) {
	// Unset any defaults, this only affects in memory state. Any Env Vars will
	// remain intact on disk
	err := os.Unsetenv(siaWalletPassword)
	if err != nil {
		t.Error(err)
	}

	// Test Default Wallet Password
	pw := WalletPassword()
	if pw != "" {
		t.Errorf("Expected wallet password to be blank but was %v", pw)
	}

	// Test Env Variable
	newPW := "abc123"
	err = os.Setenv(siaWalletPassword, newPW)
	if err != nil {
		t.Error(err)
	}
	pw = WalletPassword()
	if pw != newPW {
		t.Errorf("Expected wallet password to be %v but was %v", newPW, pw)
	}
}

// TestTurtleDexExchangeRate tests getting and setting the TurtleDex Exchange Rate
func TestTurtleDexExchangeRate(t *testing.T) {
	// Unset any defaults, this only affects in memory state. Any Env Vars will
	// remain intact on disk
	err := os.Unsetenv(siaExchangeRate)
	if err != nil {
		t.Error(err)
	}

	// Test Default
	rate := ExchangeRate()
	if rate != "" {
		t.Errorf("Expected exchange rate to be blank but was %v", rate)
	}

	// Test Env Variable
	newRate := "abc123"
	err = os.Setenv(siaExchangeRate, newRate)
	if err != nil {
		t.Error(err)
	}
	rate = ExchangeRate()
	if rate != newRate {
		t.Errorf("Expected exchange rate to be %v but was %v", newRate, rate)
	}
}
