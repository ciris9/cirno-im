package serv

import (
	"bytes"
	cim "cirno-im"
	"cirno-im/constants"
	"cirno-im/container"
	"cirno-im/logger"
	"cirno-im/wire"
	"cirno-im/wire/pkt"
	"cirno-im/wire/token"
	"fmt"
	"regexp"
	"time"
)

var log = logger.WithFields(logger.Fields{
	"services": "gateway",
	"pkg":      "serv",
})

type Handler struct {
	ServiceID string
	AppSecret string
}

func (h *Handler) Accept(conn cim.Conn, timeout time.Duration) (string, cim.Meta, error) {
	log := logger.WithFields(logger.Fields{
		"ServiceID": h.ServiceID,
		"module":    "Handler",
		"handler":   "Accept",
	})
	log.Infoln("enter")
	//1.读取登录包
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return "", nil, err
	}
	frame, err := conn.ReadFrame()
	if err != nil {
		return "", nil, err
	}
	buffer := bytes.NewBuffer(frame.GetPayload())
	req, err := pkt.MustReadLogicPkt(buffer)
	if err != nil {
		return "", nil, err
	}
	//2.需要是登录包
	if req.Command != wire.CommandLoginSignIn {
		resp := pkt.NewFrom(&req.Header)
		resp.Status = pkt.Status_InvalidCommand
		err = conn.WriteFrame(cim.OpBinary, pkt.Marshal(resp))
		if err != nil {
			log.Errorln(err)
		}
		return "", nil, fmt.Errorf("must be a invalidCommand command")
	}

	//3.反序列化body
	var login pkt.LoginRequest
	err = req.ReadBody(&login)
	if err != nil {
		return "", nil, err
	}
	//4.使用默认的DefaultSecret 解析token
	tk, err := token.Parse(token.DefaultSecret, login.Token)
	if err != nil {
		resp := pkt.NewFrom(&req.Header)
		resp.Status = pkt.Status_Unauthorized
		err = conn.WriteFrame(cim.OpBinary, pkt.Marshal(resp))
		if err != nil {
			log.Errorln(err)
		}
		return "", nil, err
	}
	//5.生成全局唯一的ChannelID
	channelID := generateChannelID(h.ServiceID, tk.Account)

	req.ChannelID = channelID
	req.WriteBody(&pkt.Session{
		ChannelID: channelID,
		GateID:    h.ServiceID,
		Account:   tk.Account,
		RemoteIP:  getIP(conn.RemoteAddr().String()),
		App:       tk.App,
	})

	//6.login转发给Login服务
	err = container.Forward(wire.SNLogin, req)
	if err != nil {
		return "", nil, err
	}
	return channelID, nil, nil
}

func (h *Handler) Receive(agent cim.Agent, payload []byte) {
	buf := bytes.NewBuffer(payload)
	packet, err := pkt.Read(buf)
	if err != nil {
		log.Errorln(err)
		return
	}
	if basicPkt, ok := packet.(*pkt.BasicPkt); ok {
		if basicPkt.Code == pkt.CodePing {
			err := agent.Push(pkt.Marshal(&pkt.BasicPkt{Code: pkt.CodePong}))
			if err != nil {
				log.Errorln(err)
			}
		}
		return
	}
	if logicPkt, ok := packet.(*pkt.LogicPkt); ok {
		logicPkt.ChannelID = agent.ID()

		messageInTotal.WithLabelValues(h.ServiceID, wire.SNTGateway, logicPkt.Command).Inc()
		messageInFlowBytes.WithLabelValues(h.ServiceID, wire.SNTGateway, logicPkt.Command).Add(float64(len(payload)))

		// 把meta注入到header中
		if agent.GetMetadata() != nil {
			logicPkt.AddStringMeta(constants.MetaKeyApp, agent.GetMetadata()[constants.MetaKeyApp])
			logicPkt.AddStringMeta(constants.MetaKeyAccount, agent.GetMetadata()[constants.MetaKeyAccount])
		}

		err := container.Forward(logicPkt.ServiceName(), logicPkt)
		if err != nil {
			logger.WithFields(logger.Fields{
				"module": "handler",
				"id":     agent.ID(),
				"cmd":    logicPkt.Command,
				"dest":   logicPkt.Dest,
			}).Error(err)
		}
	}
}

func (h *Handler) DisConnect(id string) error {
	log.Infof("disconnect %s", id)

	logout := pkt.New(wire.CommandLoginSignOut, pkt.WithChannel(id))
	err := container.Forward(wire.SNLogin, logout)
	if err != nil {
		logger.WithFields(logger.Fields{
			"module": "handler",
			"id":     id,
		}).Error(err)
		return err
	}
	return nil
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
