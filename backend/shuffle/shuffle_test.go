package shuffle

import (
	"bytes"
	"compress/zlib"
	"io"
	"testing"
)

// deshuffleFixture was produced by the real SPTushonka.Server.Core 5.0.0-pre
// RequestEncryptionUtil.ShuffleInPlace, zlib-compressing and shuffling the
// payload `{"hello":"world","n":12345,"arr":[1,2,3]}`.
var deshuffleFixture = []byte{
	52, 120, 218, 178, 202, 73, 212, 0, 49, 47, 0, 205, 207, 120, 210, 142, 202, 44, 50, 178, 201,
	49, 83, 173, 213, 54, 87, 42, 49, 163, 74, 201, 0, 3, 178, 42, 218, 7, 86, 49, 201, 202, 205,
	81, 210, 171, 86, 72, 0, 82, 171, 202, 138, 5, 82, 81, 201, 81, 50, 72, 11, 54,
}

const deshuffleFixtureExpected = `{"hello":"world","n":12345,"arr":[1,2,3]}`

func TestDeshuffleMatchesRealServerOutput(t *testing.T) {
	frame := append([]byte(nil), deshuffleFixture...)
	payload, err := Deshuffle(frame)
	if err != nil {
		t.Fatalf("Deshuffle() failed: %v", err)
	}

	reader, err := zlib.NewReader(bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("zlib.NewReader() failed: %v", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("zlib decompress failed: %v", err)
	}

	if string(decompressed) != deshuffleFixtureExpected {
		t.Errorf("Deshuffle() = %q, want %q", decompressed, deshuffleFixtureExpected)
	}
}

func TestDeshuffleShortInput(t *testing.T) {
	frame := []byte{1, 2, 3}
	out, err := Deshuffle(frame)
	if err != nil {
		t.Fatalf("Deshuffle() failed: %v", err)
	}
	if !bytes.Equal(out, frame) {
		t.Errorf("Deshuffle() = %v, want passthrough %v", out, frame)
	}
}

// TestShuffleRoundTrip verifies Shuffle produces a frame Deshuffle can reverse
// back to the original payload, for a range of lengths (the permutation is
// keyed by total frame length, so this exercises different swap sequences).
//
// The GiveItem payload below was also cross-validated out-of-band against the
// real SPTushonka.Server.Core 5.0.0-pre package: feeding Shuffle's output for
// it into the real RequestEncryptionUtil.DeShuffleAsync recovers the exact
// same JSON, confirming the server actually accepts frames this package
// produces (not just that Shuffle and Deshuffle agree with each other).
func TestShuffleRoundTrip(t *testing.T) {
	payloads := []string{
		"",
		"a",
		`{"action":"GiveItem","item":"544fb45d4bdc2dee738b4568","amount":5}`,
	}
	for _, p := range payloads {
		frame := Shuffle([]byte(p))
		out, err := Deshuffle(frame)
		if err != nil {
			t.Fatalf("Deshuffle(Shuffle(%q)) failed: %v", p, err)
		}
		if string(out) != p {
			t.Errorf("Deshuffle(Shuffle(%q)) = %q", p, out)
		}
	}
}
