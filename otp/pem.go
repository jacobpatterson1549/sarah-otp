package otp

import (
	"bytes"
	"encoding/pem"
	"errors"
)

// encode encodes the byte array with PEM encoding.
func encode(b []byte) ([]byte, error) {
	var buff bytes.Buffer
	blk := pem.Block{
		Type:  "OTP",
		Bytes: b,
	}
	err := pem.Encode(&buff, &blk)
	if err != nil {
		return nil, errors.New("applying pem encoding: " + err.Error())
	}
	return buff.Bytes(), nil
}

// decode decodes the byte array with PEM encoding.
func decode(b []byte) ([]byte, error) {
	blk, rest := pem.Decode(b)
	switch {
	case blk == nil:
		return nil, errors.New("no PEM data to decode")
	case len(rest) != 0:
		return nil, errors.New("extra text after PEM data")
	default:
		return blk.Bytes, nil
	}
}
