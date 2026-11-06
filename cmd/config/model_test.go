package config

import "testing"

func TestNormalizeBackendValid(t *testing.T) {
	backend := map[string]Backend{
		"openai": {Type: "openai"},
		"ollama": {Type: "ollama"},
	}
	for _, b := range []string{"openai", "ollama"} {
		m := Model{Name: "x", Model: "y", Backend: b}.Validate(backend)
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
	backend := map[string]Backend{}
	Model{Name: "x", Model: "y", Backend: "bogus"}.Validate(backend)
}
