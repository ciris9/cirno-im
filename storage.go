package cim

import (
	"cirno-im/wire/pkt"
	"errors"
)

var _ SessionStorage

var ErrSessionNil = errors.New("err:session nil")

type SessionStorage interface {
	Add(session *pkt.Session) error
	Delete(account string, channelID string) error
	Get(channelID string) (*pkt.Session, error)
	GetLocations(account ...string) ([]*Location, error)
	GetLocation(account string, device string) (*Location, error)
}
