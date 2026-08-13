package imgkit

import (
	"errors"
	"io"
	"path/filepath"

	"github.com/kovidgoyal/imaging"
)

const ThumbnailWidth = 200

// Thumbnail 图片缩略图
func Thumbnail(w io.Writer, filename string, rect *Rect, options ...imaging.EncodeOption) error {
	if rect == nil || rect.W < 0 || rect.H < 0 {
		return errors.New("invalid param rect")
	}

	img, err := imaging.Open(filepath.Clean(filename))
	if err != nil {
		return err
	}

	size := img.Bounds().Size()
	if rect.W == 0 && rect.H == 0 {
		rect.W = ThumbnailWidth
		rect.H = rect.W * size.Y / size.X
	} else {
		if rect.W > size.X {
			rect.W = size.X
		}
		if rect.H > size.Y {
			rect.H = size.Y
		}
		if rect.W > 0 {
			if rect.H == 0 {
				rect.H = rect.W * size.Y / size.X
			}
		} else {
			rect.W = rect.H * size.X / size.Y
		}
	}

	thumbnail := imaging.Thumbnail(img, rect.W, rect.H, imaging.Lanczos)

	format, _ := imaging.FormatFromFilename(filename)
	return imaging.Encode(w, thumbnail, format, options...)
}

// ThumbnailFromReader 图片缩略图
func ThumbnailFromReader(w io.Writer, r io.Reader, format imaging.Format, rect *Rect, options ...imaging.EncodeOption) error {
	if rect == nil || rect.W < 0 || rect.H < 0 {
		return errors.New("invalid param rect")
	}

	img, err := imaging.Decode(r)
	if err != nil {
		return err
	}

	size := img.Bounds().Size()
	if rect.W == 0 && rect.H == 0 {
		rect.W = ThumbnailWidth
		rect.H = rect.W * size.Y / size.X
	} else {
		if rect.W > size.X {
			rect.W = size.X
		}
		if rect.H > size.Y {
			rect.H = size.Y
		}
		if rect.W > 0 {
			if rect.H == 0 {
				rect.H = rect.W * size.Y / size.X
			}
		} else {
			rect.W = rect.H * size.X / size.Y
		}
	}
	thumbnail := imaging.Thumbnail(img, rect.W, rect.H, imaging.Lanczos)

	return imaging.Encode(w, thumbnail, format, options...)
}
