package config

import "testing"

func TestNormalizeBackendDefault(t *testing.T) {
	m := Model{Name: "x", Model: "y"}.Normalize()
	if m.Backend != "openai" {
		t.Fatalf("expected empty backend to default to openai, got %q", m.Backend)
	}
}

func TestNormalizeBackendValid(t *testing.T) {
	for _, b := range []string{"openai", "ollama"} {
		m := Model{Name: "x", Model: "y", Backend: b}.Normalize()
		if m.Backend != b {
			t.Fatalf("expected backend %q preserved, got %q", b, m.Backend)
		}
	}
}

func TestNormalizeBackendInvalidPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for invalid backend")
		}
	}()
	Model{Name: "x", Model: "y", Backend: "bogus"}.Normalize()
}
