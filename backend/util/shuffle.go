package util

import (
	"encoding/binary"
	"fmt"
)

// deshuffleFrame reverses the byte-shuffling SPT 5.0 applies to compressed HTTP
// responses (see SPTarkov.Server.Core.Utils.RequestEncryptionUtil.ShuffleInPlace),
// returning the zlib-compressed payload the shuffle wraps.
//
// The frame is [4-byte little-endian payload length][payload][padding], and the
// whole frame is permuted with a length-derived, keyless swap sequence. The
// server reverses it by replaying the same per-index swap decisions in the
// opposite order, which undoes the permutation because each swap is its own
// inverse.
func deshuffleFrame(data []byte) ([]byte, error) {
	n := len(data)
	if n < 4 {
		return data, nil
	}

	offset := int64(n % 2731)
	for i := 1; i < n; i++ {
		j := shuffleEntry(int64(i) + offset) % uint32(i)
		data[i], data[j] = data[j], data[i]
	}

	payloadLength := int32(binary.LittleEndian.Uint32(data[0:4]))
	if payloadLength < 0 || int(payloadLength) > n-4 {
		return nil, fmt.Errorf("deshuffled body carries an impossible length (%d)", payloadLength)
	}
	return data[4 : 4+payloadLength], nil
}

// shuffleEntry is the deterministic pseudo-random function the server uses to
// pick each swap target, keyed only by position (see RequestEncryptionUtil.Entry).
func shuffleEntry(index int64) uint32 {
	v := 1624453*(417645+index) + 1023920427
	return uint32(uint64(v) % 2396496269)
}
