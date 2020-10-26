// Package otp encrypts and decrypts messages using One-Time-Pad encription.
package otp

import (
	"crypto/rand"
	"errors"
)

// MaxKeyLength is the maximum size of keys.
const MaxKeyLength = 50000

// KeyGenerator is reader that is used to generate keys.
var KeyGenerator = rand.Reader

// Encrypt encrypts the message using the key to produce the cipher text.
func Encrypt(message, key string) ([]byte, error) {
	k, err := decode([]byte(key))
	switch {
	case err != nil:
		return nil, errors.New("decoding key: " + err.Error())
	case len(message) > len(k):
		return nil, errors.New("message must not be longer than key")
	}
	m := []byte(message)
	c := xor(m, k)
	e, err := encode(c)
	if err != nil {
		return nil, errors.New("encoding encrypted message: " + err.Error())
	}
	return e, nil
}

// Decrypt decrypts the cipher text using the key to produce the message.
func Decrypt(cipher, key string) ([]byte, error) {
	c, err := decode([]byte(cipher))
	if err != nil {
		return nil, errors.New("decoding cipher text: " + err.Error())
	}
	k, err := decode([]byte(key))
	switch {
	case err != nil:
		return nil, errors.New("decoding key: " + err.Error())
	case len(c) > len(k):
		return nil, errors.New("cipher text must not be longer than key")
	}
	m := xor(c, k)
	return m, nil
}

// GenerateKey creates an encoded key that that encodes a message of up to the specified number of characters.
func GenerateKey(length int) ([]byte, error) {
	switch {
	case length <= 0:
		return nil, errors.New("key must have positive number of characters")
	case length > MaxKeyLength:
		return nil, errors.New("key length too large")
	}
	b := make([]byte, length)
	n, err := KeyGenerator.Read(b)
	switch {
	case err != nil:
		return nil, errors.New("generating key: " + err.Error())
	case n != length:
		return nil, errors.New("could not create key of desired length")
	}
	return encode(b)
}

// Xor performs the exclusize-or operation on the two arrays, returning an array the size of the largest array.
func xor(a, b []byte) []byte {
	n := len(a)
	if len(b) > n {
		n = len(b)
	}
	c := make([]byte, n)
	for i := 0; i < len(a) && i < len(b); i++ {
		c[i] = a[i] ^ b[i]
	}
	switch {
	case n > len(a):
		copy(c[len(a):], b[len(a):])
	case n > len(b):
		copy(c[len(b):], a[len(b):])
	}
	return c
}
