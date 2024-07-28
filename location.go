package cim

import (
	"bytes"
	"cirno-im/wire/endian"
	"errors"
)

type Location struct {
	ChannelID string
	GateID    string
}

func (l *Location) Bytes() []byte {
	if l == nil {
		return []byte{}
	}
	buf := new(bytes.Buffer)
	_ = endian.WriteShortBytes(buf, []byte(l.ChannelID))
	_ = endian.WriteShortBytes(buf, []byte(l.GateID))
	return buf.Bytes()
}

func (l *Location) Unmarshal(data []byte) (err error) {
	if len(data) == 0 {
		return errors.New("data is empty")
	}
	buf := bytes.NewBuffer(data)
	l.ChannelID, err = endian.ReadShortString(buf)
	if err != nil {
		return
	}
	l.GateID, err = endian.ReadShortString(buf)
	if err != nil {
		return
	}
	return nil
}
