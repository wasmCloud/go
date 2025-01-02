package wadm

import (
	"testing"
)

func TestParseJSONManifest(t *testing.T) {
	testManifestParser(t, "*.json", ParseJSONManifest)
}

func TestParseYAMLManifest(t *testing.T) {
	testManifestParser(t, "*.yaml", ParseYAMLManifest)
}

func testManifestParser(t *testing.T, pattern string, parser func([]byte) (*Manifest, error)) {
	validFixtures := listTestdata("valid", pattern)
	for _, fixture := range validFixtures {
		t.Run(fixture, func(t *testing.T) {
			rawManifest, err := loadTestdata(fixture)
			if err != nil {
				t.Fatal(err)
			}

			manifest, err := parser(rawManifest)
			if err != nil {
				t.Fatal(err)
			}

			if err := manifest.Validate(); len(err) != 0 {
				t.Errorf("invalid manifest: %s", err)
			}
		})
	}

	invalidFixtures := listTestdata("invalid", pattern)
	for _, fixture := range invalidFixtures {
		t.Run(fixture, func(t *testing.T) {
			rawManifest, err := loadTestdata(fixture)
			if err != nil {
				t.Fatal(err)
			}

			manifest, err := parser(rawManifest)
			if err == nil {
				if !manifest.IsValid() {
					// parsed manifest, but failed validation
					return
				}
				t.Error("manifest passed parsing and validation, expected error")
			}
		})
	}

}
