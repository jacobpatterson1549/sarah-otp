// +build js,wasm

package ui

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"syscall/js"
	"time"

	"github.com/jacobpatterson1549/sarah-otp/otp"
)

var (
	encryptKeyReader    js.Value
	decryptKeyReader    js.Value
	decryptCipherReader js.Value
	encryptKey          string
	decryptKey          string
	decryptCipherText   string
)

// addFileReader registers functions to the map to disable the file input and submit button until the file is read to the destination.
func addFileReader(jsFuncs map[string]js.Func, reader *js.Value, fileDestination *string, fileInputQuery, submitButtonQuery string) {
	global := js.Global()
	fileReader := global.Get("FileReader")
	*reader = fileReader.New()
	readEventsJsFunc := NewJsEventFunc(func(event js.Value) {
		eventType := event.Get("type").String()
		switch eventType {
		case "load":
			SetButtonDisabled(fileInputQuery, false)
			SetButtonDisabled(submitButtonQuery, false)
			result := reader.Get("result")
			*fileDestination = result.String()
		case "abort":
			logInfo("reading file aborted for: " + fileInputQuery)
		case "error":
			logError("error reading file: " + fileInputQuery)
		default:
			logError("unknown file read event: " + eventType)
		}
		SetButtonDisabled(fileInputQuery, false)
		SetButtonDisabled(submitButtonQuery, false)
	})
	fileInput := QuerySelector(fileInputQuery)
	inputChangeJsFunc := NewJsEventFunc(func(event js.Value) {
		files := fileInput.Get("files")
		file := files.Index(0)
		if file.Truthy() {
			SetButtonDisabled(fileInputQuery, true)
			SetButtonDisabled(submitButtonQuery, true)
			reader.Call("addEventListener", "load", readEventsJsFunc)
			reader.Call("addEventListener", "abort", readEventsJsFunc)
			reader.Call("addEventListener", "error", readEventsJsFunc)
			reader.Call("readAsText", file)
		}
	})
	fileInput.Call("addEventListener", "change", inputChangeJsFunc)
	jsFuncs[fileInputQuery+"_readEvents"] = readEventsJsFunc
	jsFuncs[fileInputQuery+"_inputChange"] = inputChangeJsFunc
}

func initOtp(ctx context.Context, wg *sync.WaitGroup) {
	jsFuncs := make(map[string]js.Func, 6)
	addFileReader(jsFuncs, &encryptKeyReader, &encryptKey, "#encrypt-key", "#encrypt-submit")
	addFileReader(jsFuncs, &decryptKeyReader, &decryptKey, "#decrypt-key", "#decrypt-submit")
	addFileReader(jsFuncs, &decryptCipherReader, &decryptCipherText, "#decrypt-cipher", "#decrypt-submit")
	wg.Add(1)
	go ReleaseJsFuncsOnDone(ctx, wg, jsFuncs)
}

// encryptMessage is executed when the user encrypts a message using a key.
func encryptMessage(event js.Value) {
	message := Value("#encrypt-message")
	cipher, err := otp.Encrypt(message, encryptKey)
	if err != nil {
		logError("could not encrypt message: " + err.Error())
		return
	}
	savePem("cipher", cipher)
}

// decryptCipher is executed when the user decrypts a cipher using a key.
func decryptCipher(event js.Value) {
	message, err := otp.Decrypt(decryptCipherText, decryptKey)
	if err != nil {
		logError("could not decrypt cipher: " + err.Error())
		return
	}
	trimmedMessage := strings.TrimRight(string(message), "\x00")
	SetValue("#decrypted-message", trimmedMessage)
	SetChecked(".has-decrypted-message", true)
}

// generateKey is executed when the user creates a new key.
func generateKey(event js.Value) {
	keySizeText := Value("#key-size")
	keySize, err := strconv.Atoi(keySizeText)
	if err != nil {
		logError("could not convert key size to number: " + err.Error())
		return
	}
	key, err := otp.GenerateKey(keySize)
	if err != nil {
		logError("could not create key file: " + err.Error())
		return
	}
	savePem("key", key)
}

// savePem creates a new timestamped pem file and downloads it through the user's browser.
func savePem(name string, data []byte) {
	time := FormatTime(time.Now().Unix())
	time = strings.ReplaceAll(time, ":", "_")
	fileName := name + "_" + time + ".pem"
	global := js.Global()
	blob := global.Get("Blob")
	dataArr := []interface{}{
		string(data),
	}
	fileType := map[string]interface{}{
		"type": "text/plain",
	}
	fileBlob := blob.New(dataArr, fileType)
	url := global.Get("URL")
	fileURL := url.Call("createObjectURL", fileBlob)
	defer url.Call("revokeObjectURL", fileURL)
	document := global.Get("document")
	a := document.Call("createElement", "a")
	a.Set("href", fileURL)
	a.Set("download", fileName)
	a.Call("click")
	logInfo("downloaded " + fileName)
}
