package config

import "testing"

func TestLoadConfig(t *testing.T) {
	expected := Configuration{
		Title: "testConfig",
		Server: Server{
			IP:      "0.0.0.0",
			Port:    "1234",
			Logfile: "testlog",
		},
	}

	config := LoadConfig("./testdata/config_test.toml")

	if config != expected {
		t.Errorf("Generated config is different. Expected: %+v, got: %+v", expected, config)
	}
}

func TestLoadConfigDefault(t *testing.T) {
	expected := NewDefaultConfig()

	config := LoadConfig("no_file")

	if config != expected {
		t.Errorf("Generated config is different. Expected: %+v, got: %+v", expected, config)
	}
}
