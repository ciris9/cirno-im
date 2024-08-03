package handler

import (
	cim "cirno-im"
	"cirno-im/logger"
	"cirno-im/wire/pkt"
	"errors"
)

type LoginHandler struct{}

func NewLoginHandler() *LoginHandler {
	return &LoginHandler{}
}

func (h *LoginHandler) DoSyncLogin(ctx cim.Context) {

	//1.序列化
	var session pkt.Session
	if err := ctx.ReadBody(&session); err != nil {
		responseWithError(ctx, pkt.Status_InvalidPacketBody, err)
		logger.Errorln("1. serialization,err:", err)
		return
	}
	logger.WithFields(logger.Fields{
		"Func":      "Login",
		"ChannelId": session.GetChannelID(),
		"Account":   session.GetAccount(),
		"RemoteIP":  session.GetRemoteIP(),
	}).Infof("do login")

	//2.查看账号是否已经登录在其他的地方
	location, err := ctx.GetLocation(session.Account, "")
	if err != nil && !errors.Is(err, cim.ErrSessionNil) {
		responseWithError(ctx, pkt.Status_SystemException, err)
		logger.Errorln("location,err:", err)
		return
	}
	if location != nil {
		//3.通知这个用户下线
		err := ctx.Dispatch(&pkt.KickOutNotify{ChannelID: location.ChannelID})
		if err != nil {
			responseWithError(ctx, pkt.Status_SystemException, err)
			logger.Errorln("dispatch,err:", err)
			return
		}
	}

	//4.添加到会话管理器内
	if err := ctx.Add(&session); err != nil {
		responseWithError(ctx, pkt.Status_SystemException, err)
		logger.Errorln("4.add err:", err)
		return
	}

	//5.返回一个登陆成功的信息
	var resp = &pkt.LoginResponse{ChannelID: session.GetChannelID()}
	err = ctx.Resp(pkt.Status_Success, resp)
	if err != nil {
		responseWithError(ctx, pkt.Status_SystemException, err)
		logger.Errorln("5. response success,err:", err)
		return
	}
}

func (h *LoginHandler) DoSysLogout(ctx cim.Context) {
	logger.WithFields(logger.Fields{
		"Func":      "Logout",
		"ChannelId": ctx.Session().GetChannelID(),
		"Account":   ctx.Session().GetAccount(),
	}).Info("do Logout ")
	err := ctx.Delete(ctx.Session().GetAccount(), ctx.Session().GetChannelID())
	if err != nil {
		responseWithError(ctx, pkt.Status_SystemException, err)
		return
	}

	err = ctx.Resp(pkt.Status_Success, nil)
	if err != nil {
		responseWithError(ctx, pkt.Status_SystemException, err)
		return
	}
}

func responseWithError(ctx cim.Context, status pkt.Status, err error) {
	err1 := ctx.RespWithError(status, err)
	if err1 != nil {
		logger.Errorln("response with error:", err1)
	}
}
