// Copyright 2026 Tamer Badawy
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package main

import (
	"context"
	"fmt"
	"slices"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/tamer-badawy/OmniGo/ai"
	"github.com/tamer-badawy/OmniGo/assets"
	"github.com/tamer-badawy/OmniGo/ui"
)

type OmniApp struct {
	Window           fyne.Window
	Preferences      fyne.Preferences
	context          context.Context
	ActiveAI         ai.AIClient
	ActiveAIProvider AIProviderInfo
	StorageManager   *StorageManager

	SessionList    []*ChatSession
	CurrentSession *ChatSession
	SidebarList    *widget.List

	AISelector *widget.Select

	ChatLogContainer *fyne.Container
	ScrollContainer  *container.Scroll
	PromptField      *widget.Entry
}

func main() {

	// Create a new Fyne application
	myApp := app.NewWithID("com.tamerbadawy.omnigo")
	myWindow := myApp.NewWindow("OmniGo -Unified AI Client")

	if assets.ResourceLogoPng != nil {
		myWindow.SetIcon(assets.ResourceLogoPng)
	}

	omniApp := &OmniApp{
		Window:           myWindow,
		Preferences:      myApp.Preferences(),
		context:          context.Background(),
		ChatLogContainer: container.NewVBox(),
	}

	//Create Sidebar

	//Create Sidebar header
	newChatButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		initialSession := &ChatSession{
			ID:        fmt.Sprintf("session_%d", time.Now().UnixNano()),
			Messages:  []ai.Message{},
			Title:     "New Conversation Thread b",
			UpdatedAt: time.Now(),
		}
		omniApp.SessionList = append([]*ChatSession{initialSession}, omniApp.SessionList...)
		omniApp.CurrentSession = initialSession
		omniApp.SidebarList.Refresh()
		omniApp.SidebarList.Select(0)
		omniApp.ChatLogContainer.Objects = nil
		omniApp.ChatLogContainer.Refresh()
	})
	newChatButton.Importance = widget.LowImportance
	sidebarHeader := container.NewHBox(
		widget.NewLabelWithStyle("Conversations", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		newChatButton,
	)

	// Create Sidebar content
	omniApp.SidebarList = widget.NewList(
		func() int {
			return len(omniApp.SessionList)
		},
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("")
			lbl.Truncation = fyne.TextTruncateEllipsis
			delButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			delButton.Importance = widget.LowImportance
			return container.NewBorder(nil, nil, nil, delButton, lbl)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			container := o.(*fyne.Container)

			delButton := container.Objects[1].(*widget.Button)
			delButton.OnTapped = func() {
				if omniApp.SessionList[i] == omniApp.CurrentSession {
					omniApp.SidebarList.Unselect(i)
					omniApp.CurrentSession = nil
				}
				omniApp.SessionList = append(omniApp.SessionList[:i], omniApp.SessionList[i+1:]...)
				omniApp.SidebarList.Refresh()
			}
			lbl := container.Objects[0].(*widget.Label)
			lbl.SetText(omniApp.SessionList[i].Title)
			lbl.Truncation = fyne.TextTruncateEllipsis
		},
	)
	omniApp.SidebarList.OnSelected = func(id widget.ListItemID) {
		omniApp.SwitchChatSession(id)
	}

	imgLogo := canvas.NewImageFromResource(assets.ResourceLogoPng)
	imgLogo.SetMinSize(fyne.NewSize(150, 150))
	imgLogo.FillMode = canvas.ImageFillContain

	sidebar := container.NewBorder(sidebarHeader, imgLogo, nil, nil, container.NewVScroll(omniApp.SidebarList))

	// Create top bar

	omniApp.AISelector = widget.NewSelect(GetAIProvidersName(), func(value string) {
		omniApp.Preferences.SetString("AIProvider", value)
		selectedProviderIndex := slices.IndexFunc(AIProviders, func(p AIProviderInfo) bool {
			return p.Name == value
		})
		omniApp.ActiveAIProvider = AIProviders[selectedProviderIndex]
		err := omniApp.SetActiveAIClient()
		if err != nil {

			ui.ShowSettingDialog(omniApp.Preferences, omniApp.Window, omniApp.LoadSavedAIProvider)

		}
	})
	omniApp.LoadSavedAIProvider()

	storage, err := NewStorageManager(omniApp.ActiveAIProvider.Name)
	if err != nil {
		panic(err) // Crach safty catch on boot if OS system directory failed
	}

	omniApp.StorageManager = storage

	settingButton := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		ui.ShowSettingDialog(omniApp.Preferences, omniApp.Window, omniApp.LoadSavedAIProvider)
	})
	settingButton.Importance = widget.LowImportance
	topBar := container.NewHBox(
		widget.NewLabel("Active AI:"),
		omniApp.AISelector,
		layout.NewSpacer(),
		settingButton,
	)

	// Create right panel
	omniApp.ScrollContainer = container.NewVScroll(omniApp.ChatLogContainer)

	omniApp.PromptField = widget.NewMultiLineEntry()
	omniApp.PromptField.SetPlaceHolder("Ask anything across models...")
	omniApp.PromptField.OnSubmitted = func(s string) {
		go omniApp.HandlePromptSubmission()
	}

	previewPanel := ui.NewPreviewPanel([]ui.PreviewPanelItem{})

	attachButton := widget.NewButtonWithIcon("", theme.MailAttachmentIcon(), func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err == nil && reader != nil {
				// Handle file selection
				previewPanel.AddItem(ui.PreviewPanelItem{
					Title:        reader.URI().Name(),
					ThumbnailURI: reader.URI(),
				})
			}
		}, myWindow)
		fileDialog.Show()
	})
	sendButton := widget.NewButtonWithIcon("", theme.MailSendIcon(), func() {
		go omniApp.HandlePromptSubmission()
	})

	inputBar := container.NewBorder(nil, nil, attachButton, sendButton, omniApp.PromptField)

	bottomPanel := container.NewVBox(previewPanel, inputBar)

	rightPanel := container.NewBorder(topBar, bottomPanel, nil, nil, omniApp.ScrollContainer)

	split := container.NewHSplit(sidebar, rightPanel)
	split.Offset = 0.25

	myWindow.SetContent(split)
	myWindow.Resize(fyne.NewSize(950, 650))
	omniApp.LoadAndSynchronizeHistoryOnLaunch()
	myWindow.ShowAndRun()

}
