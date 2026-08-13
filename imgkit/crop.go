package imgkit

import (
	"errors"
	"image"
	"io"
	"path/filepath"

	"github.com/kovidgoyal/imaging"
)

// Crop 图片裁切
func Crop(w io.Writer, filename string, rect *Rect, options ...imaging.EncodeOption) error {
	if rect == nil || rect.X < 0 || rect.Y < 0 || rect.W <= 0 || rect.H <= 0 {
		return errors.New("invalid param rect")
	}

	img, err := imaging.Open(filepath.Clean(filename))
	if err != nil {
		return err
	}

	crop := imaging.Crop(img, image.Rect(rect.X, rect.Y, rect.X+rect.W, rect.Y+rect.H))

	format, _ := imaging.FormatFromFilename(filename)
	return imaging.Encode(w, crop, format, options...)
}

// CropFromReader 图片裁切
func CropFromReader(w io.Writer, r io.Reader, format imaging.Format, rect *Rect, options ...imaging.EncodeOption) error {
	if rect == nil || rect.X < 0 || rect.Y < 0 || rect.W < 0 || rect.H < 0 {
		return errors.New("invalid param rect")
	}

	img, err := imaging.Decode(r)
	if err != nil {
		return err
	}

	crop := imaging.Crop(img, image.Rect(rect.X, rect.Y, rect.X+rect.W, rect.Y+rect.H))

	return imaging.Encode(w, crop, format, options...)
}
