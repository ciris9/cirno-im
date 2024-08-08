package handler

import (
	cim "cirno-im"
	"cirno-im/wire/pkt"
	"github.com/pkg/errors"
	"log"
)

func responseWithError(ctx cim.Context, status pkt.Status, err error) {
	err1 := ctx.RespWithError(status, err)
	if err1 != nil {
		err2 := errors.Wrap(err1, "responseWithError Err")
		if err2 != nil {
			log.Println("responseWithError err:", err2)
		}
	}
}
