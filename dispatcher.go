package cim

import "cirno-im/wire/pkt"

type Dispatcher interface {
	Push(gateway string, channels []string, p *pkt.LogicPkt) error
}
