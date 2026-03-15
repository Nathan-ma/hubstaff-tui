package config

import "testing"

func TestBellEnabled_DefaultTrue(t *testing.T) {
	cfg := UIConfig{}
	if !cfg.BellEnabled() {
		t.Fatal("expected BellEnabled to default to true when Bell is nil")
	}
}

func TestBellEnabled_ExplicitTrue(t *testing.T) {
	v := true
	cfg := UIConfig{Bell: &v}
	if !cfg.BellEnabled() {
		t.Fatal("expected BellEnabled to be true")
	}
}

func TestBellEnabled_ExplicitFalse(t *testing.T) {
	v := false
	cfg := UIConfig{Bell: &v}
	if cfg.BellEnabled() {
		t.Fatal("expected BellEnabled to be false")
	}
}
