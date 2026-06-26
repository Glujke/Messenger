package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// tappablePreview shows an image and handles click to open a larger view.
type tappablePreview struct {
	widget.BaseWidget
	image   *canvas.Image
	onTap   func()
	visible bool
}

func newTappablePreview() *tappablePreview {
	p := &tappablePreview{
		image: canvas.NewImageFromResource(nil),
	}
	p.image.FillMode = canvas.ImageFillContain
	p.ExtendBaseWidget(p)
	return p
}

func (p *tappablePreview) setResource(res fyne.Resource) {
	p.image.Resource = res
	p.image.Refresh()
}

func (p *tappablePreview) hidePreview() {
	p.visible = false
	p.image.Hide()
	p.image.Resource = nil
	p.image.SetMinSize(fyne.NewSize(0, 0))
	p.onTap = nil
	p.Refresh()
}

func (p *tappablePreview) showPreview(size fyne.Size, onTap func()) {
	p.visible = true
	p.onTap = onTap
	p.image.Show()
	p.image.SetMinSize(size)
	p.Refresh()
}

func (p *tappablePreview) MinSize() fyne.Size {
	if !p.visible {
		return fyne.NewSize(0, 0)
	}
	return p.image.MinSize()
}

func (p *tappablePreview) Tapped(*fyne.PointEvent) {
	if p.onTap != nil {
		p.onTap()
	}
}

func (p *tappablePreview) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(p.image)
}
