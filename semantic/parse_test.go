package semantic_test

import (
	"errors"
	"testing"

	"github.com/google/osv-scalibr/semantic"
)

// knownEcosystems returns a list of ecosystems that `lockfile` supports
// automatically inferring an extractor for based on a file path.
func knownEcosystems() []string {
	return []string{
		"npm",
		"NuGet",
		"crates.io",
		"RubyGems",
		"Packagist",
		"Go",
		"Hex",
		"Maven",
		"PyPI",
		"Pub",
		"ConanCenter",
		"CRAN",
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	ecosystems := knownEcosystems()

	ecosystems = append(ecosystems, "Alpine", "Debian", "Ubuntu")

	for _, ecosystem := range ecosystems {
		_, err := semantic.Parse("", ecosystem)

		if errors.Is(err, semantic.ErrUnsupportedEcosystem) {
			t.Errorf("'%s' is not a supported ecosystem", ecosystem)
		}
	}
}

func TestMustParse(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic - '%s'", r)
		}
	}()

	ecosystems := knownEcosystems()

	ecosystems = append(ecosystems, "Alpine", "Debian", "Ubuntu")

	for _, ecosystem := range ecosystems {
		semantic.MustParse("", ecosystem)
	}
}

func TestMustParse_Panic(t *testing.T) {
	t.Parallel()

	defer func() { _ = recover() }()

	semantic.MustParse("", "<unknown>")

	// if we reached here, then we can't have panicked
	t.Errorf("function did not panic when given an unknown ecosystem")
}