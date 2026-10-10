// Package shuffle implements SPT 5.0's HTTP body framing: a length-prefixed
// frame that's permuted byte-for-byte before being sent, and unpermuted on
// arrival (see SPTarkov.Server.Core.Utils.RequestEncryptionUtil). The server
// applies this to both directions on any route other than /launcher,
// /client/metadata, /v2/shop and /files: it shuffles every compressed GET
// response, and it unconditionally expects every POST/PUT body to arrive
// pre-shuffled, compressed or not.
//
// The permutation has no secret key - it's a deterministic, length-keyed swap
// sequence, reversed by replaying the same per-index swap decisions in the
// opposite order (each swap is its own inverse).
package shuffle

import (
	"encoding/binary"
	"fmt"
)

// Deshuffle reverses the permutation and returns the frame's payload (the
// bytes between the 4-byte length header and any trailing padding).
func Deshuffle(data []byte) ([]byte, error) {
	n := len(data)
	if n < 4 {
		return data, nil
	}

	offset := int64(n % 2731)
	for i := 1; i < n; i++ {
		j := entry(int64(i) + offset) % uint32(i)
		data[i], data[j] = data[j], data[i]
	}

	payloadLength := int32(binary.LittleEndian.Uint32(data[0:4]))
	if payloadLength < 0 || int(payloadLength) > n-4 {
		return nil, fmt.Errorf("deshuffled body carries an impossible length (%d)", payloadLength)
	}
	return data[4 : 4+payloadLength], nil
}

// Shuffle wraps payload in a [4-byte length][payload] frame (no padding) and
// applies the permutation, producing a frame Deshuffle can reverse.
func Shuffle(payload []byte) []byte {
	n := len(payload)
	frame := make([]byte, 4+n)
	binary.LittleEndian.PutUint32(frame[0:4], uint32(n))
	copy(frame[4:], payload)

	offset := int64(len(frame) % 2731)
	for i := len(frame) - 1; i > 0; i-- {
		j := entry(int64(i) + offset) % uint32(i)
		frame[i], frame[j] = frame[j], frame[i]
	}
	return frame
}

// entry is the deterministic pseudo-random function the server uses to pick
// each swap target, keyed only by position (see RequestEncryptionUtil.Entry).
func entry(index int64) uint32 {
	v := 1624453*(417645+index) + 1023920427
	return uint32(uint64(v) % 2396496269)
}
