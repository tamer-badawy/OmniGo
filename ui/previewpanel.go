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

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type PreviewPanelItem struct {
	Title        string
	ThumbnailURI fyne.URI
}

type PreviewPanel struct {
	widget.BaseWidget
	Items         []PreviewPanelItem
	ThumbnailList *fyne.Container
}

func NewPreviewPanel(items []PreviewPanelItem) *PreviewPanel {
	p := &PreviewPanel{
		Items: items,
	}
	p.ExtendBaseWidget(p)
	return p
}

func (p *PreviewPanel) AddItem(item PreviewPanelItem) {
	p.Items = append(p.Items, item)
	p.Refresh()
}

func (p *PreviewPanel) RemoveItem(index int) {
	if index >= 0 && index < len(p.Items) {
		p.Items = append(p.Items[:index], p.Items[index+1:]...)
		p.Refresh()
	}
}

func (p *PreviewPanel) CreateRenderer() fyne.WidgetRenderer {
	// Create a list to display the preview items
	previewBackground := canvas.NewRectangle(color.NRGBA{R: 53, G: 114, B: 214, A: 35})

	p.ThumbnailList = container.NewHBox()

	c := container.NewStack(previewBackground, container.NewPadded(p.ThumbnailList))

	return widget.NewSimpleRenderer(c)
}

func createImgPreview(uri fyne.URI, size fyne.Size) fyne.CanvasObject {
	if uri.Extension() == ".png" || uri.Extension() == ".jpg" || uri.Extension() == ".jpeg" || uri.Extension() == ".gif" {
		img := canvas.NewImageFromURI(uri)
		img.SetMinSize(size)
		img.FillMode = canvas.ImageFillContain
		return img
	}
	return container.NewGridWrap(size, widget.NewFileIcon(uri))
}

func (p *PreviewPanel) Refresh() {
	p.ThumbnailList.Objects = nil // Clear existing thumbnails

	for index, item := range p.Items {
		thumbnail := createImgPreview(item.ThumbnailURI, fyne.NewSize(50, 50))
		thumbLabel := widget.NewLabel(item.Title)
		thumbLabel.Truncation = fyne.TextTruncateEllipsis

		thumbClearBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
			p.RemoveItem(index)

		})
		thumbClearBtn.Importance = widget.LowImportance

		thumbnailContainer := container.NewBorder(nil, thumbLabel, thumbnail, thumbClearBtn, nil)
		p.ThumbnailList.Add(thumbnailContainer)
	}

	p.ThumbnailList.Refresh() // Refresh the list to show updated thumbnails
	if len(p.Items) == 0 {
		p.Hide() // Hide the panel if there are no items
	} else {
		p.Show() // Show the panel if there are items
	}

}
