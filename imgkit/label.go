package imgkit

import (
	"errors"
	"io"
	"path/filepath"

	"git.sr.ht/~sbinet/gg"
	"github.com/kovidgoyal/imaging"
)

// Label 图片标注
func Label(w io.Writer, filename string, rects []*Rect, options ...imaging.EncodeOption) error {
	img, err := imaging.Open(filepath.Clean(filename))
	if err != nil {
		return err
	}

	dc := gg.NewContextForImage(img)
	dc.SetRGB255(255, 0, 0)
	dc.SetLineWidth(8)
	for _, rect := range rects {
		if rect.X < 0 || rect.Y < 0 || rect.W <= 0 || rect.H <= 0 {
			return errors.New("invalid param rects")
		}
		dc.DrawRectangle(float64(rect.X), float64(rect.Y), float64(rect.W), float64(rect.H))
	}
	dc.Stroke()

	format, _ := imaging.FormatFromFilename(filename)
	return imaging.Encode(w, dc.Image(), format, options...)
}

// LabelFromReader 图片标注
func LabelFromReader(w io.Writer, r io.Reader, format imaging.Format, rects []*Rect, options ...imaging.EncodeOption) error {
	img, err := imaging.Decode(r)
	if err != nil {
		return err
	}

	dc := gg.NewContextForImage(img)
	dc.SetRGB255(255, 0, 0)
	dc.SetLineWidth(8)
	for _, rect := range rects {
		if rect.X < 0 || rect.Y < 0 || rect.W <= 0 || rect.H <= 0 {
			return errors.New("invalid param rects")
		}
		dc.DrawRectangle(float64(rect.X), float64(rect.Y), float64(rect.W), float64(rect.H))
	}
	dc.Stroke()

	return imaging.Encode(w, dc.Image(), format, options...)
}
