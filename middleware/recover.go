package middleware

import (
	"fmt"

	"runtime"
	"strings"

	"cirno-im"
	"cirno-im/logger"
	"cirno-im/wire/pkt"
)

func Recover() cim.HandlerFunc {
	return func(ctx cim.Context) {
		defer func() {
			if err := recover(); err != nil {
				var callers []string
				for i := 1; ; i++ {
					_, file, line, got := runtime.Caller(i)
					if !got {
						break
					}
					callers = append(callers, fmt.Sprintf("%s:%d", file, line))
				}
				logger.WithFields(logger.Fields{
					"ChannelId": ctx.Header().ChannelID,
					"Command":   ctx.Header().Command,
					"Seq":       ctx.Header().Sequence,
				}).Error(err, strings.Join(callers, "\n"))

				_ = ctx.Resp(pkt.Status_SystemException, &pkt.ErrorResponse{Message: "SystemException"})
			}
		}()

		ctx.Next()
	}

}
