package otp

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestEncrypt(t *testing.T) {
	encryptTests := []struct {
		message string
		key     string
		want    string
		wantOk  bool
	}{
		{
			message: "CAT",
			// 12345 :
			key: `-----BEGIN OTP-----
AQIDBAU=
-----END OTP-----
`,
			// BCW45 :
			want: `-----BEGIN OTP-----
QkNXBAU=
-----END OTP-----
`,
			wantOk: true,
		},
		{
			message: "BET",
			// 06345 :
			key: `-----BEGIN OTP-----
AAYDBAU=
-----END OTP-----
`,
			// same ciphertext as previous case
			want: `-----BEGIN OTP-----
QkNXBAU=
-----END OTP-----
`,
			wantOk: true,
		},
		{ // no key
		},
		{ // message longer than key
			message: "CATASTROPHE",
			// 12345 :
			key: `-----BEGIN OTP-----
AQIDBAU=
-----END OTP-----
`,
		},
	}
	for i, test := range encryptTests {
		got, err := Encrypt(test.message, test.key)
		switch {
		case !test.wantOk:
			if err == nil {
				t.Errorf("test %v: wanted error", i)
			}
		case err != nil:
			t.Errorf("test %v: unwanted error: %v", i, err)
		case !reflect.DeepEqual([]byte(test.want), got):
			t.Errorf("test %v: not equal\nwanted: %v\ngot:    %v", i, test.want, string(got))
		}
	}
}

func TestDecrypt(t *testing.T) {
	decryptTests := []struct {
		want   string
		key    string
		cipher string
		wantOk bool
	}{
		{
			want: "CAT\x00\x00",
			// 12345 :
			key: `-----BEGIN OTP-----
AQIDBAU=
-----END OTP-----
`,
			// BCW45 :
			cipher: `-----BEGIN OTP-----
QkNXBAU=
-----END OTP-----
`,
			wantOk: true,
		},
		{
			want: "BET\x00\x00",
			// 06345 :
			key: `-----BEGIN OTP-----
AAYDBAU=
-----END OTP-----
`,
			// same ciphertext as previous case
			cipher: `-----BEGIN OTP-----
QkNXBAU=
-----END OTP-----
`,
			wantOk: true,
		},
		{ // no cipher
		},
		{ // no key
			cipher: `-----BEGIN OTP-----
QkNXBAU=
-----END OTP-----
`,
		},
		{ // cipher longer than key
			// 12345 :
			key: `-----BEGIN OTP-----
AQIDBAU=
-----END OTP-----
`,
			cipher: `-----BEGIN OTP-----
Q0FTVFJPUEhF
-----END OTP-----
`,
		},
	}
	for i, test := range decryptTests {
		got, err := Decrypt(test.cipher, test.key)
		switch {
		case !test.wantOk:
			if err == nil {
				t.Errorf("test %v: wanted error", i)
			}
		case err != nil:
			t.Errorf("test %v: unwanted error: %v", i, err)
		case !reflect.DeepEqual([]byte(test.want), got):
			t.Errorf("test %v: not equal\nwanted: %v\ngot:    %v", i, test.want, got)
		}
	}
}

func TestGenerateKey(t *testing.T) {
	generateKeyTests := []struct {
		keyLength    int
		keyGenerator io.Reader
		want         string
		wantOk       bool
	}{
		{
			keyLength:    5,
			keyGenerator: strings.NewReader(string([]byte{1, 2, 3, 4, 5})),
			want: `-----BEGIN OTP-----
AQIDBAU=
-----END OTP-----
`,
			wantOk: true,
		},
		{
			keyLength:    3,
			keyGenerator: strings.NewReader("12345"),
			want: `-----BEGIN OTP-----
MTIz
-----END OTP-----
`,
			wantOk: true,
		},
		{
			keyLength: -1,
		},
		{
			keyLength: 0,
		},
		{
			keyLength: 1234567890,
		},
		{
			keyLength:    10,
			keyGenerator: strings.NewReader("12345"),
		},
		{
			keyLength:    10,
			keyGenerator: &errorReader{},
		},
	}
	for i, test := range generateKeyTests {
		KeyGenerator = test.keyGenerator
		got, err := GenerateKey(test.keyLength)
		switch {
		case !test.wantOk:
			if err == nil {
				t.Errorf("test %v: wanted error", i)
			}
		case err != nil:
			t.Errorf("test %v: unwanted error: %v", i, err)
		case test.want != string(got):
			t.Errorf("test %v: not equal\nwanted: %v\ngot:    %v", i, test.want, string(got))
		}
	}
}

func TestXor(t *testing.T) {
	xorTests := []struct {
		a    []byte
		b    []byte
		want []byte
	}{
		{
			a:    []byte{0x0f, 0xff, 0x02},
			b:    []byte{0xf0, 0x00, 0x02, 0x03, 0x04},
			want: []byte{0xff, 0xff, 0x00, 0x03, 0x04},
		},
		{
			a:    []byte{0xf0, 0x00, 0x02, 0x03, 0x04},
			b:    []byte{0x0f, 0xff, 0x02},
			want: []byte{0xff, 0xff, 0x00, 0x03, 0x04},
		},
		{
			a:    []byte{0x01, 12},
			b:    []byte{0x02, 10},
			want: []byte{0x03, 6},
		},
		{
			b:    []byte{0xff},
			want: []byte{0xff},
		},
		{
			want: []byte{},
		},
	}
	for i, test := range xorTests {
		got := xor(test.a, test.b)
		if !reflect.DeepEqual(test.want, got) {
			t.Errorf("test %v: not equal\nwanted: %v\ngot:    %v", i, test.want, got)
		}
	}
}

type errorReader struct{}

func (r *errorReader) Read(b []byte) (n int, err error) {
	return len(b), errors.New("errorReader returns an error when read")
}
