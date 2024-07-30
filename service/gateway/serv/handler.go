package serv

import (
	cim "cirno-im"
	"cirno-im/logger"
	"cirno-im/wire"
	"fmt"
	"regexp"
	"time"
)

var log = logger.WithFields(logger.Fields{
	"service": "gateway",
	"pkg":     "serv",
})

type Handler struct {
	ServiceID string
}

func (h *Handler) Accept(conn cim.Conn, timeout time.Duration) (string, error) {

}

func (h *Handler) Receive(agent cim.Agent, payload []byte) {

}

func (h *Handler) DisConnect(id string) error {

}

var ipExp = regexp.MustCompile(string("\\:[0-9]+$"))

func getIP(remoteAddr string) string {
	if remoteAddr == "" {
		return ""
	}
	return ipExp.ReplaceAllString(remoteAddr, "")
}

func generateChannelID(serviceID, account string) string {
	return fmt.Sprintf("%s_%s_%d", serviceID, account, wire.Seq.Next())
}
