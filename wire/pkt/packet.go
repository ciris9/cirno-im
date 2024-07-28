package pkt

import (
	"cirno-im/wire"
	"cirno-im/wire/endian"
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/proto"
	"io"
	"strconv"
	"strings"
)

// LogicPkt 是网关对外部client的消息结构
type LogicPkt struct {
	Header
	Body []byte
}

type HeaderOption func(header *Header)

func WithStatus(status Status) HeaderOption {
	return func(header *Header) {
		header.Status = status
	}
}

func WithSequence(seq uint32) HeaderOption {
	return func(header *Header) {
		header.Sequence = seq
	}
}

func WithChannel(channelID string) HeaderOption {
	return func(h *Header) {
		h.ChannelID = channelID
	}
}

func WithDest(dest string) HeaderOption {
	return func(h *Header) {
		h.Dest = dest
	}
}

func New(command string, options ...HeaderOption) *LogicPkt {
	pkt := &LogicPkt{}
	pkt.Command = command
	for _, option := range options {
		option(&pkt.Header)
	}
	if pkt.Sequence == 0 {
		pkt.Sequence = wire.Seq.Next()
	}
	return pkt
}

func NewFrom(header *Header) *LogicPkt {
	pkt := &LogicPkt{}
	pkt.Header = Header{
		Command:   header.Command,
		ChannelID: header.ChannelID,
		Sequence:  header.Sequence,
		Status:    header.Status,
		Dest:      header.Dest,
	}
	return pkt
}

func (p *LogicPkt) Decode(r io.Reader) error {
	headerBytes, err := endian.ReadBytes(r)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(headerBytes, &p.Header); err != nil {
		return err
	}
	p.Body, err = endian.ReadBytes(r)
	if err != nil {
		return err
	}
	return nil
}

func (p *LogicPkt) Encode(w io.Writer) error {
	headerBytes, err := proto.Marshal(&p.Header)
	if err != nil {
		return err
	}
	if err := endian.WriteBytes(w, headerBytes); err != nil {
		return err
	}
	if err := endian.WriteBytes(w, p.Body); err != nil {
		return err
	}
	return nil
}

func (p *LogicPkt) ReadBody(val proto.Message) error {
	return json.Unmarshal(p.Body, val)
}

func (p *LogicPkt) WriteBody(val proto.Message) *LogicPkt {
	if val == nil {
		return p
	}
	p.Body, _ = json.Marshal(val)
	return p
}

func (p *LogicPkt) StringBody() string {
	return string(p.Body)
}

func (p *LogicPkt) String() string {
	return fmt.Sprintf("header:%v body:%dbits", &p.Header, len(p.Body))
}

func (p *LogicPkt) ServiceName() string {
	arr := strings.SplitN(p.Command, ".", 2)
	if len(arr) <= 1 {
		return "default"
	}
	return arr[0]
}

func (p *LogicPkt) AddMeta(m ...*Meta) {
	p.Meta = append(p.Meta, m...)
}

func (p *LogicPkt) AddStringMeta(key, value string) {
	p.AddMeta(&Meta{
		Key:   key,
		Value: value,
		Type:  MetaType_string,
	})
}

func (p *LogicPkt) GetMeta(key string) (any, bool) {
	for _, meta := range p.Meta {
		if meta.Key == key {
			switch meta.Type {
			case MetaType_int:
				v, err := strconv.Atoi(meta.Value)
				if err != nil {
					return nil, false
				}
				return v, true
			case MetaType_float:
				v, err := strconv.ParseFloat(meta.Value, 64)
				if err != nil {
					return nil, false
				}
				return v, true
			}
			return meta.Value, true
		}
	}
	return nil, false
}

func (p *LogicPkt) DelMeta(key string) {
	for i, m := range p.Meta {
		if m.Key == key {
			length := len(p.Meta)
			if i < length-1 {
				copy(p.Meta[i:], p.Meta[i+1:])
			}
			p.Meta = p.Meta[:length-1]
		}
	}
}
