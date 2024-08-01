package serv

import (
	"bytes"
	cim "cirno-im"
	"cirno-im/container"
	"cirno-im/logger"
	"cirno-im/wire"
	"cirno-im/wire/pkt"
	"errors"
	"google.golang.org/protobuf/proto"
	"strings"
	"time"
)

var log = logger.WithFields(logger.Fields{
	"service": wire.SNChat,
	"pkg":     "serv",
})

type ServerDispatcher struct {
}

func (d *ServerDispatcher) Push(gateway string, channels []string, p *pkt.LogicPkt) error {
	p.AddStringMeta(wire.MetaDestChannels, strings.Join(channels, ","))
	return container.Push(gateway, p)
}

// Disconnect default listener
func (h *ServHandler) Disconnect(id string) error {
	logger.Warnf("close event of %s", id)
	return nil
}

type ServHandler struct {
	r          *cim.Router
	cache      cim.SessionStorage
	dispatcher cim.Dispatcher
}

func NewServHandler(r *cim.Router, cache cim.SessionStorage) *ServHandler {
	return &ServHandler{
		r:          r,
		cache:      cache,
		dispatcher: &ServerDispatcher{},
	}
}

func (h *ServHandler) Accept(conn cim.Conn, timeout time.Duration) (string, error) {
	log.Infoln("enter")
	err := conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return "", err
	}
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", err
	}
	var req pkt.InnerHandshakeRequest
	if err := proto.Unmarshal(frame.GetPayload(), &req); err != nil {
		return "", err
	}
	log.Info("Accept -- ", req.ServiceID)
	return req.ServiceID, nil
}

func (h *ServHandler) Receive(agent cim.Agent, payload []byte) {
	buf := bytes.NewBuffer(payload)
	packet, err := pkt.MustReadLogicPkt(buf)
	if err != nil {
		log.Errorln(err)
		return
	}
	var session *pkt.Session
	if packet.Command == wire.CommandLoginSignIn {
		server, _ := packet.GetMeta(wire.MetaDestServer)
		session = &pkt.Session{
			ChannelID: packet.ChannelID,
			GateID:    server.(string),
			Tags:      []string{"AuthGenerated"},
		}
	} else {
		session, err = h.cache.Get(packet.ChannelID)
		if errors.Is(err, cim.ErrSessionNil) {
			RespErr(agent, packet, pkt.Status_SessionNotFound)
			return
		} else if err != nil {
			RespErr(agent, packet, pkt.Status_SystemException)
			return
		}
	}
	log.Debugf("recv a message from %s  %s", session, &packet.Header)
	if err := h.r.Serve(packet, h.dispatcher, h.cache, session); err != nil {
		log.Warn(err)
	}
}

func RespErr(ag cim.Agent, p *pkt.LogicPkt, status pkt.Status) {
	packet := pkt.NewFrom(&p.Header)
	packet.Status = status
	packet.Flag = pkt.Flag_Response
	err := ag.Push(pkt.Marshal(packet))
	if err != nil {
		log.Errorln(err)
	}
	return
}
