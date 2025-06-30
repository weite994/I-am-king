package ssecmd

import (
	"testing"
)

func TestNewServer(t *testing.T) {
	// Test that a new server can be created with a valid config
	config := Config{
		Token:   "test-token",
		Version: "test-version",
		Address: "localhost:8080",
	}

	server, err := NewServer(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if server == nil {
		t.Fatal("Expected non-nil server")
	}

	// Test that server creation fails with an empty token
	invalidConfig := Config{
		// No token set
		Address: "localhost:8080",
	}

	_, err = NewServer(invalidConfig)
	if err == nil {
		t.Fatal("Expected error for missing token, got nil")
	}
}

func TestCreateServerWithOptions(t *testing.T) {
	// Test that a server can be created with functional options
	server, err := CreateServerWithOptions(
		WithToken("test-token"),
		WithAddress("localhost:9090"),
		WithVersion("1.0.0"),
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if server == nil {
		t.Fatal("Expected non-nil server")
	}

	if server.config.Address != "localhost:9090" {
		t.Errorf("Expected address to be 'localhost:9090', got %s", server.config.Address)
	}

	if server.config.Version != "1.0.0" {
		t.Errorf("Expected version to be '1.0.0', got %s", server.config.Version)
	}
}

func TestDefaultConfig(t *testing.T) {
	// Test default configuration values
	config := DefaultConfig()

	if config.Address != "localhost:8080" {
		t.Errorf("Expected default address to be 'localhost:8080', got %s", config.Address)
	}

	if config.BasePath != "" {
		t.Errorf("Expected default base path to be empty, got %s", config.BasePath)
	}

	if config.ReadOnly != false {
		t.Error("Expected default ReadOnly to be false")
	}
}
