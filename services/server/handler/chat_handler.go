package handler

import (
	cim "cirno-im"
	"cirno-im/services/server/service"
	"cirno-im/wire/pkt"
	"cirno-im/wire/rpc"
	"errors"
	"time"
)

var ErrNoDestination = errors.New("dest is empty")

type ChatHandler struct {
	msgService   service.Message
	groupService service.Group
}

func NewChatHandler(message service.Message, group service.Group) *ChatHandler {
	return &ChatHandler{
		msgService:   message,
		groupService: group,
	}
}

func (h *ChatHandler) DoUserTalk(ctx cim.Context) {
	if ctx.Header().Dest == "" {
		responseWithError(ctx, pkt.Status_NoDestination, ErrNoDestination)
		return
	}

	//解包
	var req pkt.MessageRequest
	if err := ctx.ReadBody(&req); err != nil {
		responseWithError(ctx, pkt.Status_InvalidPacketBody, err)
		return
	}

	//获取接收方的位置信息
	receiver := ctx.Header().GetDest()
	location, err := ctx.GetLocation(receiver, "")
	if err != nil && !errors.Is(err, cim.ErrSessionNil) {
		responseWithError(ctx, pkt.Status_InvalidPacketBody, err)
		return
	}

	//保存离线消息
	sendTime := time.Now().UnixNano()
	response, err := h.msgService.InsertUser(ctx.Session().GetApp(), &rpc.InsertMessageReq{
		Sender:   ctx.Session().GetAccount(),
		Dest:     receiver,
		SendTime: sendTime,
		Message: &rpc.Message{
			Type:  req.GetType(),
			Body:  req.GetBody(),
			Extra: req.GetExtra(),
		},
	})
	if err != nil {
		responseWithError(ctx, pkt.Status_SystemException, err)
		return
	}
	msgId := response.MessageId

	//如果接收方在线就推送一条消息过去
	if location != nil {
		if err = ctx.Dispatch(&pkt.MessagePush{
			MessageID: msgId,
			Type:      req.GetType(),
			Body:      req.GetBody(),
			Extra:     req.GetExtra(),
			Sender:    ctx.Session().GetAccount(),
			SendTime:  sendTime,
		}, location); err != nil {
			responseWithError(ctx, pkt.Status_SystemException, err)
			return
		}
	}

	//返回一条resp消息
	err = ctx.Resp(pkt.Status_Success, &pkt.MessageResponse{
		MessageID: msgId,
		SendTime:  sendTime,
	})
	if err != nil {
		responseWithError(ctx, pkt.Status_SystemException, err)
		return
	}
}

func (h *ChatHandler) DoGroupTalk(ctx cim.Context) {
	if ctx.Header().Dest == "" {
		responseWithError(ctx, pkt.Status_NoDestination, ErrNoDestination)
		return
	}
	//解包
	var req pkt.MessageRequest
	if err := ctx.ReadBody(&req); err != nil {
		responseWithError(ctx, pkt.Status_InvalidPacketBody, err)
		return
	}

	group := ctx.Header().GetDest()
	sendTime := time.Now().UnixNano()

	//保存离线消息
	resp, err := h.msgService.InsertGroup(ctx.Session().GetApp(), &rpc.InsertMessageReq{
		Sender:   ctx.Session().GetAccount(),
		Dest:     group,
		SendTime: sendTime,
		Message: &rpc.Message{
			Type:  req.GetType(),
			Body:  req.GetBody(),
			Extra: req.GetExtra(),
		},
	})
	if err != nil {
		responseWithError(ctx, pkt.Status_SystemException, err)
		return
	}

	//读取成员列表
	membersResp, err := h.groupService.Members(ctx.Session().GetApp(), &rpc.GroupMembersReq{
		GroupId: group,
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	var members = make([]string, len(membersResp.Users))
	for i, user := range membersResp.Users {
		members[i] = user.Account
	}

	//  批量寻址（群成员）
	locs, err := ctx.GetLocations(members...)
	if err != nil && !errors.Is(err, cim.ErrSessionNil) {
		responseWithError(ctx, pkt.Status_SystemException, err)
		return
	}

	//  批量推送消息给成员
	if len(locs) > 0 {
		if err = ctx.Dispatch(&pkt.MessagePush{
			MessageID: resp.MessageId,
			Type:      req.GetType(),
			Body:      req.GetBody(),
			Extra:     req.GetExtra(),
			Sender:    ctx.Session().GetAccount(),
			SendTime:  sendTime,
		}, locs...); err != nil {
			responseWithError(ctx, pkt.Status_SystemException, err)
			return
		}
	}
	//  返回一条resp消息
	_ = ctx.Resp(pkt.Status_Success, &pkt.MessageResponse{
		MessageID: resp.MessageId,
		SendTime:  sendTime,
	})
}

func (h *ChatHandler) DoTalkAck(ctx cim.Context) {
	var req pkt.MessageAckRequest
	if err := ctx.ReadBody(&req); err != nil {
		responseWithError(ctx, pkt.Status_InvalidPacketBody, err)
		return
	}
	err := h.msgService.SetAck(ctx.Session().GetApp(), &rpc.AckMessageReq{
		Account:   ctx.Session().GetAccount(),
		MessageId: req.GetMessageID(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}
	_ = ctx.Resp(pkt.Status_Success, nil)
}
