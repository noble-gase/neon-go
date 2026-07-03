package httpzip

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"io"
)

type ZipCloser struct {
	io.Reader
	rc   io.Closer
	body io.Closer
}

func (zc *ZipCloser) Close() error {
	if err := zc.rc.Close(); err != nil {
		_ = zc.body.Close()
		return err
	}
	return zc.body.Close()
}

// boundedReader 限制可读取的解压字节数；超出 Central Directory 声明的
// 未压缩大小时返回错误，防止 zip bomb 膨胀到远超声明的大小
type boundedReader struct {
	r io.Reader
	n uint64 // 剩余可读字节数
}

func (b *boundedReader) Read(p []byte) (int, error) {
	if b.n == 0 {
		// 允许正常读到 EOF
		var tmp [1]byte
		n, err := b.r.Read(tmp[:])
		if n > 0 {
			return 0, fmt.Errorf("httpzip: decompressed size exceeds declared uncompressed size")
		}
		return 0, err
	}
	if uint64(len(p)) > b.n {
		p = p[:b.n]
	}
	n, err := b.r.Read(p)
	b.n -= uint64(n)
	return n, err
}

// File 表示 ZIP 文件中的一个文件条目（Central Directory 中的记录）
type File struct {
	// Name 文件名（相对路径），来源于 Central Directory/File Header
	Name string

	// CompressedSize 压缩后的大小（字节数）
	// 对应 Central Directory 中的 compressed size 字段
	CompressedSize uint64

	// UncompressedSize 解压后的大小（字节数）
	// 对应 Central Directory 中的 uncompressed size 字段
	UncompressedSize uint64

	// Compression 压缩算法标识
	// 0 = Store（无压缩），8 = Deflate，其它见 ZIP 规范
	Compression uint16

	// Offset 文件数据在 ZIP 中的偏移量（相对于整个 ZIP 文件的开头）
	// 一般指向对应 Local File Header 的起始位置
	Offset uint64

	// reader 指向 ZIP Reader，用于按需加载该文件的数据
	// 通常实现为一个 ReaderAt + 解压逻辑
	reader *Reader
}

// Open 打开文件内容，返回一个 io.ReadCloser
func (f *File) Open() (io.ReadCloser, error) {
	// 校验偏移量边界
	if f.Offset+30 > uint64(f.reader.Size) {
		return nil, fmt.Errorf("invalid local header offset: %d", f.Offset)
	}

	// 读取 Local Header
	localHeader, err := f.reader.httpRange(int64(f.Offset), int64(f.Offset)+30+256)
	if err != nil {
		return nil, err
	}
	// Local File Header 固定部分30字节，签名 0x04034b50
	if len(localHeader) < 30 || binary.LittleEndian.Uint32(localHeader) != 0x04034b50 {
		return nil, fmt.Errorf("invalid local file header (name=%s)", f.Name)
	}

	nameLen := binary.LittleEndian.Uint16(localHeader[26:])
	extraLen := binary.LittleEndian.Uint16(localHeader[28:])

	// 文件数据起始位置 = 偏移量 + header大小 + 文件名 + extra
	dataOffset := f.Offset + 30 + uint64(nameLen) + uint64(extraLen)
	if dataOffset > uint64(f.reader.Size) || f.CompressedSize > uint64(f.reader.Size)-dataOffset {
		return nil, fmt.Errorf("file data out of bounds (name=%s)", f.Name)
	}

	// 先校验压缩方式，避免对不支持的条目发起无谓请求
	switch f.Compression {
	case 0, 8:
	default:
		return nil, fmt.Errorf("unsupported compression: %d", f.Compression)
	}

	// 零长度条目（如目录、空文件）无需发起请求
	if f.CompressedSize == 0 {
		if f.Compression == 8 && f.UncompressedSize != 0 {
			return nil, fmt.Errorf("invalid empty deflate stream (name=%s)", f.Name)
		}
		return io.NopCloser(bytes.NewReader(nil)), nil
	}

	// 读取文件数据
	compData, err := f.reader.httpRangeRaw(int64(dataOffset), int64(dataOffset)+int64(f.CompressedSize)-1)
	if err != nil {
		return nil, err
	}

	switch f.Compression {
	case 0: // Store（无压缩）
		return compData.Body, nil
	default: // Deflate
		rc := flate.NewReader(compData.Body)
		return &ZipCloser{
			// 限制解压大小，防止 zip bomb
			Reader: &boundedReader{r: rc, n: f.UncompressedSize},
			rc:     rc,
			body:   compData.Body,
		}, nil
	}
}
