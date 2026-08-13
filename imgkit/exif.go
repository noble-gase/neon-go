package imgkit

import (
	"bytes"
	"image"
	_ "image/gif" // 注册GIF解码器，用于获取图片尺寸（imagemeta 不支持GIF）
	"io"
	"os"

	"github.com/bep/imagemeta"
	"github.com/shopspring/decimal"
	_ "golang.org/x/image/bmp" // 注册BMP解码器，用于获取图片尺寸（imagemeta 不支持BMP）
)

// Rect 定义一个矩形框
type Rect struct {
	X int `json:"x"` // 左上角X坐标
	Y int `json:"y"` // 左上角Y坐标
	W int `json:"w"` // 宽度
	H int `json:"h"` // 高度
}

// Orientation 图片的旋转方向
type Orientation int

func (o Orientation) String() string {
	s := ""
	switch o {
	case TopLeft:
		s = "Top-Left"
	case TopRight:
		s = "Top-Right"
	case BottomRight:
		s = "Bottom-Right"
	case BottomLeft:
		s = "Bottom-Left"
	case LeftTop:
		s = "Left-Top"
	case RightTop:
		s = "Right-Top"
	case RightBottom:
		s = "Right-Bottom"
	case LeftBottom:
		s = "Left-Bottom"
	}
	return s
}

const (
	TopLeft     Orientation = 1
	TopRight    Orientation = 2
	BottomRight Orientation = 3
	BottomLeft  Orientation = 4
	LeftTop     Orientation = 5
	RightTop    Orientation = 6
	RightBottom Orientation = 7
	LeftBottom  Orientation = 8
)

// EXIF 定义图片EXIF
type EXIF struct {
	Size        int64
	Format      string
	Width       int
	Height      int
	Orientation string
	Longitude   decimal.Decimal
	Latitude    decimal.Decimal
}

// ParseEXIF 解析图片EXIF
func ParseEXIF(filename string) (*EXIF, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	data := &EXIF{
		Size: stat.Size(),
	}

	// 图片格式（按魔数探测，与文件扩展名无关）
	format, name := sniffFormat(f)
	if len(name) == 0 {
		return data, nil
	}
	data.Format = name

	// 解析EXIF（尽力而为，解析失败时忽略错误）
	if format != imagemeta.ImageFormatAuto {
		var tags imagemeta.Tags
		res, _ := imagemeta.Decode(imagemeta.Options{
			R:           f,
			ImageFormat: format,
			Sources:     imagemeta.EXIF | imagemeta.CONFIG,
			HandleTag: func(ti imagemeta.TagInfo) error {
				tags.Add(ti)
				return nil
			},
		})

		// 实际宽高（来自文件头，无需解码像素）
		data.Width = res.ImageConfig.Width
		data.Height = res.ImageConfig.Height

		exifTags := tags.EXIF()
		// 经纬度
		lat, lng, _ := tags.GetLatLong()
		data.Longitude = decimal.NewFromFloat(lng)
		data.Latitude = decimal.NewFromFloat(lat)
		// 转向
		if tag, ok := exifTags["Orientation"]; ok {
			data.Orientation = Orientation(tagInt(tag.Value)).String()
		}
		// 宽高缺失时，使用EXIF记录值
		if data.Width == 0 || data.Height == 0 {
			if tag, ok := exifTags["ExifImageWidth"]; ok {
				data.Width = tagInt(tag.Value)
			}
			if tag, ok := exifTags["ExifImageHeight"]; ok {
				data.Height = tagInt(tag.Value)
			}
		}
	}

	// 如果宽度或高度为0，则从图片头中获取
	if data.Width == 0 || data.Height == 0 {
		if _, err := f.Seek(0, io.SeekStart); err == nil {
			if cfg, _, err := image.DecodeConfig(f); err == nil {
				data.Width = cfg.Width
				data.Height = cfg.Height
			}
		}
	}

	return data, nil
}

// sniffFormat 通过文件魔数探测图片格式；
// imagemeta 不支持的格式（如GIF、BMP）返回 ImageFormatAuto 和格式名称，无法识别的格式返回空名称。
func sniffFormat(r io.ReaderAt) (imagemeta.ImageFormat, string) {
	buf := make([]byte, 12)
	if n, _ := r.ReadAt(buf, 0); n < 12 {
		return imagemeta.ImageFormatAuto, ""
	}

	switch {
	case bytes.HasPrefix(buf, []byte{0xFF, 0xD8, 0xFF}):
		return imagemeta.JPEG, "JPEG"
	case bytes.HasPrefix(buf, []byte("\x89PNG\r\n\x1a\n")):
		return imagemeta.PNG, "PNG"
	case bytes.HasPrefix(buf, []byte("II*\x00")), bytes.HasPrefix(buf, []byte("MM\x00*")):
		return imagemeta.TIFF, "TIFF"
	case bytes.HasPrefix(buf, []byte("RIFF")) && bytes.Equal(buf[8:12], []byte("WEBP")):
		return imagemeta.WebP, "WEBP"
	case bytes.Equal(buf[4:8], []byte("ftyp")):
		// ISO-BMFF 容器，按 major brand 区分
		switch string(buf[8:12]) {
		case "avif", "avis":
			return imagemeta.AVIF, "AVIF"
		case "heic", "heix", "hevc", "hevx":
			return imagemeta.HEIF, "HEIC"
		case "heif", "heim", "heis", "mif1", "msf1":
			return imagemeta.HEIF, "HEIF"
		}
	case bytes.HasPrefix(buf, []byte("GIF8")):
		return imagemeta.ImageFormatAuto, "GIF"
	case bytes.HasPrefix(buf, []byte("BM")):
		return imagemeta.ImageFormatAuto, "BMP"
	}
	return imagemeta.ImageFormatAuto, ""
}

// tagInt 将EXIF整型标签值转换为int（imagemeta 对 SHORT/LONG 分别返回 uint16/uint32）
func tagInt(v any) int {
	switch n := v.(type) {
	case uint16:
		return int(n)
	case uint32:
		return int(n)
	}
	return 0
}
