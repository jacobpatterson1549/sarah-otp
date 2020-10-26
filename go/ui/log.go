// +build js,wasm

package ui

import "time"

// logInfo logs an info-styled message.
func logInfo(text string) {
	addLog("info", text)
}

// logError logs an error-styled message.
func logError(text string) {
	addLog("error", text)
}

// clearLog clears the log.
func clearLog() {
	SetChecked(".has-log", false)
	logScrollElement := QuerySelector(".log>.scroll")
	logScrollElement.Set("innerHTML", "")
}

// addLog writes a log item with the specified class.
func addLog(class, text string) {
	SetChecked(".has-log", true)
	clone := CloneElement(".log>template")
	cloneChildren := clone.Get("children")
	logItemElement := cloneChildren.Index(0)
	time := FormatTime(time.Now().UTC().Unix())
	textContent := time + " : " + text
	logItemElement.Set("textContent", textContent)
	logItemElement.Set("className", class)
	logScrollElement := QuerySelector(".log>.scroll")
	logScrollElement.Call("appendChild", logItemElement)
	scrollHeight := logScrollElement.Get("scrollHeight")
	clientHeight := logScrollElement.Get("clientHeight")
	scrollTop := scrollHeight.Int() - clientHeight.Int()
	logScrollElement.Set("scrollTop", scrollTop)
}
