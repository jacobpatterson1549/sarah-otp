// +build js,wasm

package ui

import (
	"context"
	"sync"
	"syscall/js"
)

// InitDomFuncs registers the logging and otp javascript functions.
func InitDomFuncs(ctx context.Context, wg *sync.WaitGroup) {
	logFuncs := map[string]js.Func{
		"clear": NewJsFunc(clearLog),
	}
	otpFuncs := map[string]js.Func{
		"encrypt":      NewJsEventFunc(encryptMessage),
		"decrypt":      NewJsEventFunc(decryptCipher),
		"generateKey": NewJsEventFunc(generateKey),
	}
	RegisterFuncs(ctx, wg, "log", logFuncs)
	RegisterFuncs(ctx, wg, "otp", otpFuncs)
	initOtp(ctx, wg)
}

// NewJsFunc creates a new javascript function from the provided function.
func NewJsFunc(fn func()) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer AlertOnPanic()
		fn()
		return nil
	})
}

// NewJsEventFunc creates a new javascript function from the provided function that processes an event and returns nothing.
// PreventDefault is called on the event before applying the function.
func NewJsEventFunc(fn func(event js.Value)) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		event := args[0]
		event.Call("preventDefault")
		defer AlertOnPanic()
		fn(event)
		return nil
	})
}

// RegisterFuncs sets the function as fields on the parent.
// The parent object is created if it does not exist.
func RegisterFuncs(ctx context.Context, wg *sync.WaitGroup, parentName string, jsFuncs map[string]js.Func) {
	global := js.Global()
	parent := global.Get(parentName)
	if parent.IsUndefined() {
		parent = js.ValueOf(make(map[string]interface{}))
		global.Set(parentName, parent)
	}
	for fnName, fn := range jsFuncs {
		parent.Set(fnName, fn)
	}
	wg.Add(1)
	go ReleaseJsFuncsOnDone(ctx, wg, jsFuncs)
}

// ReleaseJsFuncsOnDone releases the jsFuncs and decrements the waitgroup when the context is done.
// This function should be called on a separate goroutine.
func ReleaseJsFuncsOnDone(ctx context.Context, wg *sync.WaitGroup, jsFuncs map[string]js.Func) {
	defer AlertOnPanic()
	<-ctx.Done() // BLOCKING
	for _, f := range jsFuncs {
		f.Release()
	}
	wg.Done()
}
