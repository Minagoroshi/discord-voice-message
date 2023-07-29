package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"io/fs"
	"log"
	"os"
	"strconv"
)

var token []byte

func init() {
	var err error
	token, err = os.ReadFile("token")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			_, err := os.Create("token")
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			log.Fatalln("Couldn't read token:", err)
		}
	}
}

func main() {

	a := app.NewWithID("discord-voice-message")
	w := a.NewWindow("Discord Voice Message")
	w.Resize(fyne.NewSize(600, 400))

	// Create the widgets
	channelEntry := widget.NewEntry()
	channelEntry.SetPlaceHolder("Enter channel ID")

	tokenEntry := widget.NewPasswordEntry()
	saveTkn := widget.NewButton("Save", func() {
		err := os.WriteFile("token", []byte(tokenEntry.Text), 0755)
		if err != nil {
			dialog.ShowError(err, w)
		} else {
			dialog.ShowInformation("Token Saved", "Token saved to file", w)
		}
	})

	tokenEntry.Text = string(token)
	tokenEntry.Refresh()

	audioFileEntry := widget.NewEntry()
	audioFileEntry.Disable()
	audioFileEntry.SetPlaceHolder("Select audio file...")

	browseButton := widget.NewButton("Browse", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				audioFileEntry.SetText(reader.URI().Path())
			}
		}, w)

	})

	timesEntry := widget.NewEntry()
	timesEntry.SetPlaceHolder("Number of times to send")

	sendButton := widget.NewButton("Send", func() {
		// Get the channel ID
		channel := channelEntry.Text

		// Get the audio file path
		audioFilePath := audioFileEntry.Text

		// Get the number of times to send
		x, err := strconv.Atoi(timesEntry.Text)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid number of times"), w)
			return
		}

		// show a loading dialog
		loadingDialog := dialog.NewProgress("Processing Files", "Please wait...", w)
		loadingDialog.Show()
		defer loadingDialog.Hide()

		totalFiles := float64(x)
		progressStep := 1.0 / totalFiles
		currentProgress := 0.0

		for i := 0; i < x; i++ {
			file, err := NewFile(audioFilePath)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			fmt.Printf("[Sending %d/%d] -> File Name: %s | File Type: %s | File Size: %d bytes\n", i+1, x, file.FileName, file.FileType, file.FileSize)

			_, err = file.CreateFile(tokenEntry.Text, channel)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			err = file.PutFileData()
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			err = file.SendFile(tokenEntry.Text, channel)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			fmt.Printf("[Sent %d/%d] -> File Name: %s | File Type: %s | File Size: %d bytes\n", i+1, x, file.FileName, file.FileType, file.FileSize)

			// Update the progress bar
			currentProgress += progressStep
			log.Println(currentProgress)
			loadingDialog.SetValue(currentProgress)
			loadingDialog.Refresh()
		}

		dialog.ShowInformation("Sent", "Voice messages sent successfully!", w)
	})

	// Create a container for the widgets
	mainVbox := container.NewVBox(container.NewAdaptiveGrid(2,
		container.NewVBox(
			widget.NewCard("Channel ID", "enter the channel id", channelEntry),
			widget.NewCard("Token", "enter a a discord token", container.NewVBox(tokenEntry, saveTkn))),
		container.NewVBox(
			widget.NewCard("Audio File", "select your audio file", container.NewVBox(audioFileEntry, browseButton)),
			widget.NewCard("Times to Send", "how many times do you want to send the file", timesEntry)),
	),
		sendButton)

	w.SetContent(mainVbox)
	w.ShowAndRun()
}
