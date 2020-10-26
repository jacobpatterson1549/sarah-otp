package otp

import "testing"

func TestEncode(t *testing.T) {
	b := []byte("HELLO")
	got, err := encode(b)
	want := `-----BEGIN OTP-----
SEVMTE8=
-----END OTP-----
`
	switch {
	case err != nil:
		t.Errorf("unwanted error: %v", err)
	case want != string(got):
		t.Errorf("not equal\nwanted: %v\ngot:    %v", want, string(got))
	}
}

func TestDecode(t *testing.T) {
	decodeTests := []struct {
		b      string
		want   string
		wantOk bool
	}{
		{
			b: `-----BEGIN OTP-----
SEVMTE8=
-----END OTP-----
`,
			want:   "HELLO",
			wantOk: true,
		},
		{
			b: `-----BEGIN OTP-----
SEVMTE8=
-----END OTP-----
extra data which should not be in pem file
`,
		},
		{
			b: `-----BEGIN other-pem-----
SEVMTE8=
-----END other-pem-----
`,
			want:   "HELLO",
			wantOk: true,
		},
		{
			b: "",
		},
		{
			b: `---INVALID PEM---`,
		},
	}
	for i, test := range decodeTests {
		got, err := decode([]byte(test.b))
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
