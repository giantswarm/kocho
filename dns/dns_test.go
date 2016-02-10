package dns

import "testing"

func TestDefaultNamingPattern(t *testing.T) {
	entries := DefaultNamingPattern.GetEntries("demo-cluster")

	if entries.Catchall != "*.demo-cluster.example.com" {
		t.Errorf("expected CatchAll to match string, but got: %s", entries.Catchall)
	}

	if entries.CatchallPrivate != "*.demo-cluster.private.example.com" {
		t.Errorf("expected CatchAllPrivate to match string, but got: %s", entries.CatchallPrivate)
	}

	if entries.Public != "demo-cluster.example.com" {
		t.Errorf("expected Public to match string, but got: %s", entries.Public)
	}

	if entries.Private != "demo-cluster.private.example.com" {
		t.Errorf("expected Private to match string, but got: %s", entries.Private)
	}

	if entries.Fleet != "demo-cluster.fleet.example.com" {
		t.Errorf("expected Fleet to match string, but got: %s", entries.Fleet)
	}
}
