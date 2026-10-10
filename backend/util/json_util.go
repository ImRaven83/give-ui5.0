package util

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"io"
	"spt-give-ui/backend/http"
	"spt-give-ui/backend/shuffle"
)

func GetJson(url string, sessionId string, target interface{}) error {
	data, err := GetRawBytes(url, sessionId)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func GetRawBytes(url string, sessionId string) ([]byte, error) {
	r, err := http.DoGetCompressed(url, sessionId)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	shuffled, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	frame, err := shuffle.Deshuffle(shuffled)
	if err != nil {
		return nil, err
	}

	reader, err := zlib.NewReader(bytes.NewReader(frame))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func GetRawBytesCompressed(url string, sessionId string) ([]byte, error) {
	r, err := http.DoGetCompressed(url, sessionId)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	return data, nil
}

func ParseByteResponse(profiles []byte, target interface{}) error {
	return json.Unmarshal(profiles, target)
}
