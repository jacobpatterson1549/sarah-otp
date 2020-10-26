package ui

import "testing"

func TestFormatTime(t *testing.T) {
	utcSeconds := int64(1603729564)
	want := "09:26:04"
	got := FormatTime(utcSeconds)
	if want != got {
		t.Errorf("not equal\nwanted: %v\ngot:    %v", want, got)
	}
}
