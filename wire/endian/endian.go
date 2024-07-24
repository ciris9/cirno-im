package endian

import (
	"encoding/binary"
	"io"
)

// Default 网络传输一般来讲为小端序,为了加快网络传输效率采用小端序进行读写
var Default = binary.LittleEndian

func ReadUint8(r io.Reader) (uint8, error) {
	var bytes = make([]byte, 1)
	if _, err := io.ReadFull(r, bytes); err != nil {
		return 0, err
	}
	return bytes[0], nil
}

func ReadUint16(r io.Reader) (uint16, error) {
	var bytes = make([]byte, 2)
	if _, err := io.ReadFull(r, bytes); err != nil {
		return 0, err
	}
	return Default.Uint16(bytes), nil
}

func ReadUint32(r io.Reader) (uint32, error) {
	var bytes = make([]byte, 4)
	if _, err := io.ReadFull(r, bytes); err != nil {
		return 0, err
	}
	return Default.Uint32(bytes), nil
}

func ReadUint64(r io.Reader) (uint64, error) {
	var bytes = make([]byte, 8)
	if _, err := io.ReadFull(r, bytes); err != nil {
		return 0, err
	}
	return Default.Uint64(bytes), nil
}

func ReadString(r io.Reader) (string, error) {
	bytes, err := ReadBytes(r)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func ReadBytes(r io.Reader) ([]byte, error) {
	//读取数据长度，前四位一定是数据长度
	bufLen, err := ReadUint32(r)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, bufLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func ReadFixedBytes(len int, r io.Reader) ([]byte, error) {
	buf := make([]byte, len)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func WriteUint8(w io.Writer, val uint8) error {
	buf := []byte{val}
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}

func WriteUint16(w io.Writer, val uint16) error {
	buf := make([]byte, 2)
	Default.PutUint16(buf, val)
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}

func WriteUint32(w io.Writer, val uint32) error {
	buf := make([]byte, 4)
	Default.PutUint32(buf, val)
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}

func WriteUint64(w io.Writer, val uint64) error {
	buf := make([]byte, 8)
	Default.PutUint64(buf, val)
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}

func WriteString(w io.Writer, str string) error {
	if err := WriteBytes(w, []byte(str)); err != nil {
		return err
	}
	return nil
}

func WriteBytes(w io.Writer, buf []byte) error {
	bufLen := len(buf)

	if err := WriteUint32(w, uint32(bufLen)); err != nil {
		return err
	}
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}

func WriteShortBytes(w io.Writer, buf []byte) error {
	bufLen := len(buf)

	if err := WriteUint16(w, uint16(bufLen)); err != nil {
		return err
	}
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}

func ReadShortBytes(r io.Reader) ([]byte, error) {
	bufLen, err := ReadUint16(r)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, bufLen)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func ReadShortString(r io.Reader) (string, error) {
	buf, err := ReadShortBytes(r)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
