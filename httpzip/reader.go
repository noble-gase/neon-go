package httpzip

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

// ErrInsecurePath 表示 ZIP 内存在不安全的文件路径（如绝对路径、包含 ".." 或反斜杠）。
// 此时 Reader 仍会返回（非 nil），调用方若确认来源可信，可忽略该错误继续使用。
var ErrInsecurePath = errors.New("httpzip: insecure file path in zip")

// Reader 远程 ZIP Reader
type Reader struct {
	url string

	Size int64
	File []*File

	ctx context.Context
}

// OpenReader 打开远程 ZIP 并解析目录
//
//	EOCD(End of Central Directory)
//
//	标准ZIP
//	---------------------------------------------------------
//	4   bytes   End of central dir signature (0x06054b50)
//	2   bytes   Number of this disk
//	2   bytes   Number of the disk with the start of the central directory
//	2   bytes   Total number of entries in the central directory on this disk
//	2   bytes   Total number of entries in the central directory
//	4   bytes   Size of the central directory
//	4   bytes   Offset of start of central directory with respect to the starting disk number
//	2   bytes   ZIP file comment length
//
//	ZIP64
//	---------------------------------------------------------
//	4   bytes   Signature 0x06064b50
//	8   bytes   Size of zip64 end of central directory record
//	2   bytes   Version made by
//	2   bytes   Version needed to extract
//	4   bytes   Number of this disk
//	4   bytes   Number of the disk with the start of the central directory
//	8   bytes   Total number of entries in the central directory on this disk
//	8   bytes   Total number of entries in the central directory
//	8   bytes   Size of the central directory
//	8   bytes   Offset of start of central directory
func NewReader(ctx context.Context, url string) (*Reader, error) {
	r := &Reader{
		url: url,
		ctx: ctx,
	}

	// Step 1: 获取远程文件大小（通过 HEAD 请求）
	if err := r.contentLength(); err != nil {
		return nil, err
	}

	// Step 2: 定位 EOCD (End of Central Directory)
	// ZIP 的目录信息在文件末尾，最大偏移范围 64KB
	eocdSize := min(r.Size, int64(64<<10))
	tail, err := r.httpRange(r.Size-eocdSize, r.Size-1)
	if err != nil {
		return nil, err
	}

	// EOCD signature = 0x06054b50 (小端存储：50 4b 05 06)
	sig := []byte{0x50, 0x4b, 0x05, 0x06}
	idx := strings.LastIndex(string(tail), string(sig))
	if idx < 0 {
		return nil, fmt.Errorf("EOCD not found")
	}
	eocd := tail[idx:]
	// EOCD 固定部分至少22字节
	if len(eocd) < 22 {
		return nil, fmt.Errorf("truncated EOCD")
	}

	// 从 EOCD 里解析 Central Directory 的大小和偏移量
	cdSize := uint64(binary.LittleEndian.Uint32(eocd[12:]))
	cdOffset := uint64(binary.LittleEndian.Uint32(eocd[16:]))

	// Step 3: ZIP64 兼容处理
	// 如果 cdSize 或 cdOffset == 0xFFFFFFFF，说明需要使用 ZIP64 EOCD 结构
	if cdSize == 0xFFFFFFFF || cdOffset == 0xFFFFFFFF {
		// ZIP64 EOCD Locator 签名 = 0x07064b50
		locatorSig := []byte{0x50, 0x4b, 0x06, 0x07}
		locIdx := strings.LastIndex(string(tail), string(locatorSig))
		if locIdx < 0 {
			return nil, fmt.Errorf("ZIP64 locator not found")
		}
		loc := tail[locIdx:]
		// ZIP64 EOCD Locator 固定20字节
		if len(loc) < 20 {
			return nil, fmt.Errorf("truncated ZIP64 locator")
		}

		// 读取 ZIP64 EOCD 偏移量（存放在 Locator 中）
		zip64EOCDOffset := binary.LittleEndian.Uint64(loc[8:])
		if zip64EOCDOffset > uint64(r.Size) {
			return nil, fmt.Errorf("invalid ZIP64 EOCD offset")
		}

		// 加载 ZIP64 EOCD 结构
		zip64EOCD, _err := r.httpRange(int64(zip64EOCDOffset), int64(zip64EOCDOffset)+55)
		if _err != nil {
			return nil, _err
		}
		if len(zip64EOCD) < 56 {
			return nil, fmt.Errorf("truncated ZIP64 EOCD")
		}
		if binary.LittleEndian.Uint32(zip64EOCD) != 0x06064b50 {
			return nil, fmt.Errorf("invalid ZIP64 EOCD signature")
		}

		// 从 ZIP64 EOCD 中解析 cdSize 和 cdOffset
		cdSize = binary.LittleEndian.Uint64(zip64EOCD[40:])
		cdOffset = binary.LittleEndian.Uint64(zip64EOCD[48:])
	}

	// 校验 Central Directory 边界，防止恶意声明导致越界
	if cdSize == 0 {
		return nil, fmt.Errorf("invalid central directory size: %d", cdSize)
	}
	if cdOffset > uint64(r.Size) || cdSize > uint64(r.Size)-cdOffset {
		return nil, fmt.Errorf("central directory out of bounds")
	}

	// Step 4: 加载 Central Directory 数据 (包含每个文件的元信息：名字、大小、压缩方式、偏移量)
	cdData, err := r.httpRange(int64(cdOffset), int64(cdOffset)+int64(cdSize)-1)
	if err != nil {
		return nil, err
	}

	// Step 5: 解析 Central Directory
	if err := r.parseCentralDirectory(cdData); err != nil {
		return nil, err
	}

	// Step 6: 路径穿越防护（同 archive/zip：发现不安全路径时返回 Reader 和 ErrInsecurePath）
	for _, f := range r.File {
		if f.Name == "" {
			continue
		}
		// ZIP 规范要求使用正斜杠，包含反斜杠视为不安全
		if !filepath.IsLocal(f.Name) || strings.Contains(f.Name, `\`) {
			return r, ErrInsecurePath
		}
	}
	return r, nil
}

func (r *Reader) contentLength() error {
	resp, err := Client().R().SetContext(r.ctx).Head(r.url)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status())
	}
	if resp.RawResponse.ContentLength <= 0 {
		return fmt.Errorf("unknown content length: %d", resp.RawResponse.ContentLength)
	}
	r.Size = resp.RawResponse.ContentLength

	return nil
}

func (r *Reader) httpRange(start, end int64) ([]byte, error) {
	resp, err := Client().R().
		SetContext(r.ctx).
		SetHeader("Range", fmt.Sprintf("bytes=%d-%d", start, end)).
		Get(r.url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (r *Reader) httpRangeRaw(start, end int64) (*http.Response, error) {
	resp, err := Client().R().
		SetContext(r.ctx).
		SetHeader("Range", fmt.Sprintf("bytes=%d-%d", start, end)).
		SetDoNotParseResponse(true).
		Get(r.url)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}

func (r *Reader) parseCentralDirectory(data []byte) error {
	i := 0
	for i+4 <= len(data) {
		// 每个 Central Directory File Header 都以固定的签名开头 0x02014b50
		if binary.LittleEndian.Uint32(data[i:]) != 0x02014b50 {
			break
		}
		// Central Directory File Header 固定部分46字节
		if i+46 > len(data) {
			return fmt.Errorf("truncated central directory header")
		}

		// 压缩方式 (2 bytes)
		compMethod := binary.LittleEndian.Uint16(data[i+10:])

		// 压缩大小 (4 bytes, 如果大于 4GB 会写成 0xFFFFFFFF，需用 ZIP64 extra field 解析)
		compSize := uint64(binary.LittleEndian.Uint32(data[i+20:]))

		// 未压缩大小 (4 bytes, 同上可能需要 ZIP64)
		uncompSize := uint64(binary.LittleEndian.Uint32(data[i+24:]))

		// 文件名长度
		nameLen := int(binary.LittleEndian.Uint16(data[i+28:]))

		// Extra field 长度
		extraLen := int(binary.LittleEndian.Uint16(data[i+30:]))

		// 文件注释长度
		commentLen := int(binary.LittleEndian.Uint16(data[i+32:]))

		// 对应 Local File Header 的偏移量 (4 bytes，可能需要 ZIP64)
		localHeaderOffset := uint64(binary.LittleEndian.Uint32(data[i+42:]))

		// 校验变长部分边界，防止越界
		end := i + 46 + nameLen + extraLen + commentLen
		if end > len(data) {
			return fmt.Errorf("truncated central directory entry")
		}

		// 文件名
		name := string(data[i+46 : i+46+nameLen])

		// Extra field 数据
		extra := data[i+46+nameLen : i+46+nameLen+extraLen]

		// 如果大小或偏移量是 0xFFFFFFFF，说明需要用 ZIP64 extra field 来获取真实值
		if compSize == 0xFFFFFFFF || uncompSize == 0xFFFFFFFF || localHeaderOffset == 0xFFFFFFFF {
			j := 0
			for j+4 <= len(extra) {
				// 每个 extra field: [HeaderID(2 bytes)][DataSize(2 bytes)][Data...]
				headerID := binary.LittleEndian.Uint16(extra[j:])
				dataSize := int(binary.LittleEndian.Uint16(extra[j+2:]))

				fieldEnd := j + 4 + dataSize
				if fieldEnd > len(extra) {
					return fmt.Errorf("invalid extra field (name=%s)", name)
				}

				// 0x0001 表示 ZIP64 extended information extra field
				if headerID == 0x0001 {
					k := j + 4

					// 按顺序存放未压缩大小、压缩大小、local header 偏移量
					if uncompSize == 0xFFFFFFFF {
						if k+8 > fieldEnd {
							return fmt.Errorf("invalid ZIP64 extra field (name=%s)", name)
						}
						uncompSize = binary.LittleEndian.Uint64(extra[k:])
						k += 8
					}
					if compSize == 0xFFFFFFFF {
						if k+8 > fieldEnd {
							return fmt.Errorf("invalid ZIP64 extra field (name=%s)", name)
						}
						compSize = binary.LittleEndian.Uint64(extra[k:])
						k += 8
					}
					if localHeaderOffset == 0xFFFFFFFF {
						if k+8 > fieldEnd {
							return fmt.Errorf("invalid ZIP64 extra field (name=%s)", name)
						}
						localHeaderOffset = binary.LittleEndian.Uint64(extra[k:])
					}
				}
				j = fieldEnd // 移动到下一个 extra field
			}
		}

		i = end

		r.File = append(r.File, &File{
			Name:             name,
			CompressedSize:   compSize,
			UncompressedSize: uncompSize,
			Compression:      compMethod,
			Offset:           localHeaderOffset,
			reader:           r,
		})
	}
	return nil
}
