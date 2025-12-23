package generator_test

import (
	"regexp"
	"testing"
	"url-shortener/internal/infrastructure/generator"

	"github.com/stretchr/testify/assert"
)

func TestNewRandomShortCodeGenerator(t *testing.T) {
	gen := generator.NewRandomShortCodeGenerator()

	assert.NotNil(t, gen)
}

func TestRandomShortCodeGenerator_Generate(t *testing.T) {
	gen := generator.NewRandomShortCodeGenerator()

	// Generate multiple codes to ensure randomness
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code := gen.Generate()

		// Verify length
		assert.Equal(t, generator.ShortCodeLength, len(code), "Generated code should have correct length")

		// Verify format: should only contain alphanumeric, dash, or underscore
		validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
		assert.True(t, validPattern.MatchString(code), "Code should only contain valid characters: %s", code)

		codes[code] = true
	}

	// With 100 generations, we should have some uniqueness (though collisions are possible)
	// At minimum, we should have generated at least 50 unique codes
	assert.Greater(t, len(codes), 50, "Should generate mostly unique codes")
}

func TestSanitizeCode(t *testing.T) {
	gen := generator.NewRandomShortCodeGenerator()

	// Generate many codes to test sanitization
	// Since sanitizeCode is private, we test it indirectly through Generate
	// Base64 encoding can produce characters like +, /, = which need sanitization
	for i := 0; i < 1000; i++ {
		code := gen.Generate()

		// Verify all characters are valid
		validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
		assert.True(t, validPattern.MatchString(code), "Code should be sanitized: %s", code)

		// Verify no invalid characters
		invalidPattern := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
		assert.False(t, invalidPattern.MatchString(code), "Code should not contain invalid characters: %s", code)

		// Verify length is maintained after sanitization
		assert.Equal(t, generator.ShortCodeLength, len(code), "Code length should be maintained: %s", code)
	}
}

func TestRandomShortCodeGenerator_Generate_Length(t *testing.T) {
	gen := generator.NewRandomShortCodeGenerator()

	for i := 0; i < 100; i++ {
		code := gen.Generate()
		assert.Equal(t, generator.ShortCodeLength, len(code), "All generated codes should have length %d", generator.ShortCodeLength)
	}
}
