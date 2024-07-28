package container

import (
	"cirno-im"
	"cirno-im/wire/pkt"
	"hash/crc32"
)

//HashCode generated a hash code
func HashCode(key string) (int, error) {
	ieee := crc32.NewIEEE()
	_, err := ieee.Write([]byte(key))
	if err != nil {
		return -1, err
	}
	return int(ieee.Sum32()), nil
}

// Selector is used to select a Service
type Selector interface {
	Lookup(*pkt.Header,[]cim.Service) string
}
