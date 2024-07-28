package container

import (
	"cirno-im/wire/pkt"
	"hash/crc32"
)

func HashCode(key string) (int, error) {
	ieee := crc32.NewIEEE()
	_, err := ieee.Write([]byte(key))
	if err != nil {
		return -1, err
	}
	return int(ieee.Sum32()), nil
}

type Selector interface {
	Lookup(header *pkt.Header,[]cim.)
}
