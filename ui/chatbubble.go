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
	"image/color"
	"path"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ChatBubble struct {
	widget.BaseWidget
	Role        string
	Text        string
	Attachments []string
	AIModel     string

	TextContainer *fyne.Container
}

func NewChatBubble(role, text, aiModel string, attachments ...[]string) *ChatBubble {
	var finalAttachments []string

	if len(attachments) > 0 {
		finalAttachments = attachments[0]
	}
	c := &ChatBubble{
		Role:        role,
		Text:        text,
		AIModel:     aiModel,
		Attachments: finalAttachments,
	}
	c.ExtendBaseWidget(c)
	return c
}

func generateAttachmentPreview(uri string, size fyne.Size) fyne.CanvasObject {

	if path.Ext(uri) == ".png" || path.Ext(uri) == ".jpg" || path.Ext(uri) == ".jpeg" || path.Ext(uri) == ".gif" {
		img := canvas.NewImageFromFile(uri)
		img.SetMinSize(size)
		img.FillMode = canvas.ImageFillContain
		return img
	}
	return container.NewGridWrap(size, widget.NewFileIcon(storage.NewFileURI(uri)))
}

func (c *ChatBubble) CreateRenderer() fyne.WidgetRenderer {
	var bubbleBg *canvas.Rectangle
	var senderName string

	isDark := fyne.CurrentApp().Settings().ThemeVariant() == theme.VariantDark

	if c.Role == "user" {
		if isDark {
			bubbleBg = canvas.NewRectangle(color.NRGBA{R: 41, G: 98, B: 255, A: 45}) // Blue for user in dark mode
		} else {
			bubbleBg = canvas.NewRectangle(color.NRGBA{R: 225, G: 235, B: 255, A: 255}) // Blue for user in light mode
		}

		senderName = "You"
	} else {
		if isDark {
			bubbleBg = canvas.NewRectangle(color.NRGBA{R: 255, G: 255, B: 255, A: 15}) // Light gray for AI in dark mode
		} else {
			bubbleBg = canvas.NewRectangle(color.NRGBA{R: 242, G: 242, B: 247, A: 255}) // Light gray for AI in light mode
		}
		senderName = c.AIModel
	}
	bubbleBg.CornerRadius = 8

	nameLabel := widget.NewLabelWithStyle(senderName, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	textEntry := widget.NewRichTextFromMarkdown(c.Text)
	textEntry.Wrapping = fyne.TextWrapWord

	c.TextContainer = container.NewVBox(textEntry)

	attachmentContainer := container.NewHBox()
	for _, attachment := range c.Attachments {
		thumbnail := generateAttachmentPreview(attachment, fyne.NewSize(50, 50))
		attachmentContainer.Add(thumbnail)
	}

	bubbleContent := container.NewVBox(container.NewPadded(nameLabel), c.TextContainer, attachmentContainer)

	stackContainer := container.NewStack(bubbleBg, bubbleContent)

	var aligmentRow *fyne.Container
	if c.Role == "user" {
		aligmentRow = container.NewGridWithColumns(2, layout.NewSpacer(), stackContainer)
	} else {
		aligmentRow = container.NewVBox(stackContainer)
	}

	return widget.NewSimpleRenderer(aligmentRow)

}

func (c *ChatBubble) Refresh() {
	dynamicContainer := container.NewVBox()

	segments := strings.Split(c.Text, "```")
	for i, segment := range segments {
		if strings.TrimSpace(segment) == "" {
			continue
		}

		if i%2 == 1 {
			lines := strings.SplitN(segment, "\n", 2)
			lang := strings.TrimSpace(lines[0])
			if lang == "" {
				lang = "code"
			}
			codeSnippet := ""
			if len(lines) > 1 {
				codeSnippet = lines[1]
			}
			codeBlock := c.createSnippetBlockWithCopyButton(codeSnippet, lang)
			dynamicContainer.Add(codeBlock)
		} else {

			textBlock := widget.NewRichTextFromMarkdown(segment)
			textBlock.Wrapping = fyne.TextWrapWord

			dynamicContainer.Add(textBlock)

		}
	}

	c.TextContainer.Objects = dynamicContainer.Objects
}

func (c *ChatBubble) createSnippetBlockWithCopyButton(snippetText, lang string) *fyne.Container {

	cleanCode := strings.TrimSpace(snippetText)

	// Render the snippet beautifully using RichText
	snippetRichText := widget.NewRichTextFromMarkdown("```" + lang + "\n" + cleanCode + "\n```")
	snippetRichText.Wrapping = fyne.TextWrapWord

	// Create a copy button
	var copyButton *widget.Button
	copyButton = widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		// Copy the snippet to clipboard
		fyne.CurrentApp().Clipboard().SetContent(cleanCode)
		copyButton.SetIcon(theme.ConfirmIcon())
		copyButton.Refresh()
		go func() {
			// Reset the icon after 2 seconds
			fyne.Do(func() {
				copyButton.SetIcon(theme.ContentCopyIcon())
				copyButton.Refresh()
			})
		}()

	})
	copyButton.Importance = widget.LowImportance

	buttonRow := container.NewHBox(layout.NewSpacer(), copyButton,
		widget.NewLabelWithStyle(strings.ToUpper(lang), fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
	)
	return container.NewVBox(snippetRichText, buttonRow)

}

/*
func (c *ChatBubble) createBlockquoteWithCopyButton(snippetText string) *fyne.Container {
	cleanQuote := strings.TrimSpace(snippetText)

	// Render the snippet beautifully using RichText
	snippetRichText := widget.NewRichTextFromMarkdown("> " + cleanQuote)
	snippetRichText.Wrapping = fyne.TextWrapWord

	// Create a copy button
	var copyButton *widget.Button
	copyButton = widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		// Copy the snippet to clipboard
		fyne.CurrentApp().Clipboard().SetContent(cleanQuote)
		copyButton.SetIcon(theme.ConfirmIcon())

		go func() {
			// Reset the icon after 2 seconds
			fyne.Do(func() {
				copyButton.SetIcon(theme.ContentCopyIcon())

			})
		}()

	})
	copyButton.Importance = widget.LowImportance

	buttonRow := container.NewHBox(layout.NewSpacer(), copyButton)
	return container.NewVBox(snippetRichText, buttonRow)
}
*/

func (c *ChatBubble) UpdateText(newText string) {

	c.Text = newText

	c.Refresh()
}

func (c *ChatBubble) Thinking() {
	c.UpdateText("🤖 ... Thinking")
}
