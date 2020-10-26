// +build js,wasm

package ui

import (
	"syscall/js"
	"testing"
)

func TestFormatTime(t *testing.T) {
	tzoFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return 420 // 7 hours past UTC
	})
	defer tzoFunc.Release()
	datePrototype := js.Global().Get("Date").Get("prototype")
	getTimezoneOffset := "getTimezoneOffset"
	origFunc := datePrototype.Get(getTimezoneOffset)
	defer datePrototype.Set(getTimezoneOffset, origFunc)
	datePrototype.Set(getTimezoneOffset, tzoFunc)
	utcSeconds := int64(1603729564)
	want := "09:26:04"
	got := FormatTime(utcSeconds)
	if want != got {
		t.Errorf("not equal\nwanted: %v\ngot:    %v", want, got)
	}
}
