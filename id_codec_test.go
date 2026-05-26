package instagrapi

import (
	"testing"
)

// TestInstagramIdCodec_Encode tests ID encoding to shortcodes.
func TestInstagramIdCodec_Encode(t *testing.T) {
	codec := InstagramIdCodec{}

	tests := []struct {
		name string
		num  int64
	}{
		{"zero", 0},
		{"one", 1},
		{"small", 12345},
		{"large", 9876543210},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := codec.Encode(tt.num)
			if len(result) == 0 {
				t.Errorf("Encode(%d) returned empty string", tt.num)
			}
			// Verify all characters are from the encoding set
			for _, ch := range result {
				found := false
				for i, ec := range encodingChars {
					if ch == ec {
						found = true
						t.Logf("char %c found at index %d", ch, i)
						break
					}
				}
				if !found {
					t.Errorf("Encode produced invalid character: %c", ch)
				}
			}
		})
	}

	// Test that different numbers produce different encodings
	if codec.Encode(1) == codec.Encode(2) {
		t.Error("Different IDs should produce different encodings")
	}
}

// TestInstagramIdCodec_Decode tests shortcode decoding.
func TestInstagramIdCodec_Decode(t *testing.T) {
	codec := InstagramIdCodec{}

	tests := []struct {
		name string
		num  int64
	}{
		{"zero", 0},
		{"one", 1},
		{"small", 12345},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := codec.Encode(tt.num)
			decoded := codec.Decode(encoded)
			if decoded != tt.num {
				t.Errorf("Decode(Encode(%d)) = %d; want %d", tt.num, decoded, tt.num)
			}
		})
	}

	// Test round-trip for larger numbers
	for _, num := range []int64{100, 1000, 10000, 100000, 1000000, 987654321} {
		t.Run(string(rune(num%255)), func(t *testing.T) {
			encoded := codec.Encode(num)
			decoded := codec.Decode(encoded)
			if decoded != num {
				t.Errorf("Decode(Encode(%d)) = %d; want %d", num, decoded, num)
			}
		})
	}

	// Test invalid shortcode characters are skipped
	result := codec.Decode("@#$") // all invalid chars
	if result < 0 {
		t.Errorf("Decode of invalid chars should return >= 0, got %d", result)
	}
}

// TestInstagramIdCodec_RoundTrip tests encode/decode round-trips.
func TestInstagramIdCodec_RoundTrip(t *testing.T) {
	codec := InstagramIdCodec{}

	testIDs := []int64{
		0, 1, 2, 10, 100, 1000, 10000, 100000,
		123456789, 987654321, 1234567890,
	}

	for _, id := range testIDs {
		t.Run(string(rune(id%255)), func(t *testing.T) {
			encoded := codec.Encode(id)
			decoded := codec.Decode(encoded)
			if decoded != id {
				t.Errorf("RoundTrip(%d): Encode->Decode = %s -> %d; want %d", id, encoded, decoded, id)
			}

			// Double round-trip: decode the decoded value back to string then encode
			reEncoded := codec.Encode(decoded)
			if reEncoded != encoded {
				t.Errorf("Double round-trip(%d): Encode->Decode->Encode = %s; want %s", id, reEncoded, encoded)
			}
		})
	}
}
