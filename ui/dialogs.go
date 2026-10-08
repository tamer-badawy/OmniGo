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

package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func ShowSettingDialog(preference fyne.Preferences, window fyne.Window, onClose func()) {
	geminiKeyEntry := widget.NewPasswordEntry()
	cohereKeyEntry := widget.NewPasswordEntry()
	geminiKeyEntry.SetPlaceHolder("Paste Gemini API Key")
	cohereKeyEntry.SetPlaceHolder("Paste Cohere API Key")
	geminiKeyEntry.SetText(preference.StringWithFallback("gemini_key", ""))
	cohereKeyEntry.SetText(preference.StringWithFallback("cohere_key", ""))

	descLabel := widget.NewLabel("Please provide your API keys for the AI providers you want to use.\nThese keys are necessary for authentication and access to the respective AI services.")
	descLabel.Wrapping = fyne.TextWrapWord

	form := widget.NewForm(
		widget.NewFormItem("Gemini API Key", geminiKeyEntry),
		widget.NewFormItem("Cohere API Key", cohereKeyEntry),
	)
	c := container.NewVBox(descLabel, widget.NewSeparator(), container.NewPadded(form))

	d := dialog.NewCustomConfirm("AI Settings", "Save", "Cancel", c, func(confirmed bool) {
		if confirmed {
			preference.SetString("gemini_key", geminiKeyEntry.Text)
			preference.SetString("cohere_key", cohereKeyEntry.Text)
			onClose()

		}
	}, window)
	d.Resize(fyne.NewSize(550, 300))
	d.Show()

}
