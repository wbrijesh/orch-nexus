package database

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	// Skip test if not in CI environment or if necessary env vars aren't set
	if os.Getenv("CI") != "true" || os.Getenv("KEYSPACES_ENDPOINT") == "" {
		t.Skip("Skipping Amazon Keyspaces test in local development")
	}

	srv := New()
	if srv == nil {
		t.Fatal("New() returned nil")
	}

	err := srv.Close()
	if err != nil {
		t.Fatalf("expected Close() to return nil")
	}
}

func TestHealth(t *testing.T) {
	// Skip test if not in CI environment or if necessary env vars aren't set
	if os.Getenv("CI") != "true" || os.Getenv("KEYSPACES_ENDPOINT") == "" {
		t.Skip("Skipping Amazon Keyspaces test in local development")
	}

	srv := New()

	stats := srv.Health()

	if stats["status"] != "up" {
		t.Fatalf("expected status to be up, got %s", stats["status"])
	}

	if _, ok := stats["error"]; ok {
		t.Fatalf("expected error not to be present")
	}

	if stats["message"] != "It's healthy" {
		t.Fatalf("expected message to be 'It's healthy', got %s", stats["message"])
	}

	// Check that we have some keyspaces
	if stats["keyspaces_count"] == "0" {
		t.Fatalf("expected to have at least one keyspace")
	}

	// Check that we have the AWS region
	if stats["aws_region"] == "" {
		t.Fatalf("expected aws_region to be set")
	}

	err := srv.Close()
	if err != nil {
		t.Fatalf("expected Close() to return nil")
	}
}

func TestClose(t *testing.T) {
	// Skip test if not in CI environment or if necessary env vars aren't set
	if os.Getenv("CI") != "true" || os.Getenv("KEYSPACES_ENDPOINT") == "" {
		t.Skip("Skipping Amazon Keyspaces test in local development")
	}

	srv := New()

	if srv.Close() != nil {
		t.Fatalf("expected Close() to return nil")
	}
}
