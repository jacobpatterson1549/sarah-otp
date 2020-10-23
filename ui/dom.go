// +build js,wasm

package ui

import (
	"errors"
	"strings"
	"syscall/js"
	"time"
)

// AlertOnPanic checks to see if a panic has occurred.
// Thes function shoould be deferred as the first statement for each goroutine
func AlertOnPanic() {
	if r := recover(); r != nil {
		err := recoverError(r)
		f := []string{
			"FATAL: site shutting down",
			"See browser console for more information",
			"Message: " + err.Error(),
		}
		message := strings.Join(f, "\n")
		alert(message)
		panic(err)
	}
}

// recoverError converts the recovery interface into a useful error, panicing if the interface is not an error or a string.
func recoverError(r interface{}) error {
	switch v := r.(type) {
	case error:
		return v
	case string:
		return errors.New(v)
	default:
		panic([]interface{}{"unknown panic type", v, r})
	}
}

// alert shows a popup in the browser.
func alert(message string) {
	global := js.Global()
	global.Call("alert", message)
}

// QuerySelector returns the first element returned by the query from root of the document.
func QuerySelector(query string) js.Value {
	global := js.Global()
	document := global.Get("document")
	return document.Call("querySelector", query)
}

// QuerySelectorAll returns an array of the elements returned by the query from the specified document.
func QuerySelectorAll(document js.Value, query string) []js.Value {
	value := document.Call("querySelectorAll", query)
	values := make([]js.Value, value.Length())
	for i := 0; i < len(values); i++ {
		values[i] = value.Index(i)
	}
	return values
}

// Value gets the value of the input element.
func Value(query string) string {
	element := QuerySelector(query)
	value := element.Get("value")
	return value.String()
}

// SetValue sets the value of the input element.
func SetValue(query, value string) {
	element := QuerySelector(query)
	element.Set("value", value)
}

// SetChecked sets the checked property of the element.
func SetChecked(query string, checked bool) {
	element := QuerySelector(query)
	element.Set("checked", checked)
}

// SetButtonDisabled sets the disable property of the button element.
func SetButtonDisabled(query string, disabled bool) {
	element := QuerySelector(query)
	element.Set("disabled", disabled)
}

// CloneElement creates a close of the element, which should be a template.
func CloneElement(query string) js.Value {
	templateElement := QuerySelector(query)
	contentElement := templateElement.Get("content")
	clone := contentElement.Call("cloneNode", true)
	return clone
}

// FormatTime formats a datetime to HH:MM:SS.
func FormatTime(utcSeconds int64) string {
	t := time.Unix(utcSeconds, 0) // uses local timezone
	return t.Format("15:04:05")
}
