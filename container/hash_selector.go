package container

import (
	cim "cirno-im"
	"cirno-im/logger"
	"cirno-im/wire/pkt"
)

type HashSelector struct{}


func (h *HashSelector) Lookup(header *pkt.Header, srvs []cim.Service) string {
	ll := len(srvs)
	code, err := HashCode(header.ChannelID)
	if err != nil {
		logger.Error(err.Error())
		return ""
	}
	return srvs[code%ll].ServiceID()
}
