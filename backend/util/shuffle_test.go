package util

import (
	"bytes"
	"compress/zlib"
	"io"
	"testing"
)

// shuffledFixture was produced by the real SPTushonka.Server.Core 5.0.0-pre
// RequestEncryptionUtil.ShuffleInPlace, zlib-compressing and shuffling the
// payload `{"hello":"world","n":12345,"arr":[1,2,3]}`.
var shuffledFixture = []byte{
	52, 120, 218, 178, 202, 73, 212, 0, 49, 47, 0, 205, 207, 120, 210, 142, 202, 44, 50, 178, 201,
	49, 83, 173, 213, 54, 87, 42, 49, 163, 74, 201, 0, 3, 178, 42, 218, 7, 86, 49, 201, 202, 205,
	81, 210, 171, 86, 72, 0, 82, 171, 202, 138, 5, 82, 81, 201, 81, 50, 72, 11, 54,
}

const shuffledFixtureExpected = `{"hello":"world","n":12345,"arr":[1,2,3]}`

func TestDeshuffleFrameMatchesRealServerOutput(t *testing.T) {
	frame := append([]byte(nil), shuffledFixture...)
	payload, err := deshuffleFrame(frame)
	if err != nil {
		t.Fatalf("deshuffleFrame() failed: %v", err)
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

	if string(decompressed) != shuffledFixtureExpected {
		t.Errorf("deshuffleFrame() = %q, want %q", decompressed, shuffledFixtureExpected)
	}
}

func TestDeshuffleFrameRejectsImpossibleLength(t *testing.T) {
	frame := []byte{0xFF, 0xFF, 0xFF, 0x7F, 1, 2, 3, 4}
	if _, err := deshuffleFrame(frame); err == nil {
		t.Error("deshuffleFrame() expected error for impossible length, got nil")
	}
}

func TestDeshuffleFrameShortInput(t *testing.T) {
	frame := []byte{1, 2, 3}
	out, err := deshuffleFrame(frame)
	if err != nil {
		t.Fatalf("deshuffleFrame() failed: %v", err)
	}
	if !bytes.Equal(out, frame) {
		t.Errorf("deshuffleFrame() = %v, want passthrough %v", out, frame)
	}
}
