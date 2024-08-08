package handler

import (
	cim "cirno-im"
	"cirno-im/services/server/service"
	"cirno-im/wire/pkt"
	"cirno-im/wire/rpc"
	"errors"
)

type OfflineHandler struct {
	msgService service.Message
}

func NewOfflineHandler(message service.Message) *OfflineHandler {
	return &OfflineHandler{
		msgService: message,
	}
}

func (h *OfflineHandler) DoSyncIndex(ctx cim.Context) {
	var req pkt.MessageIndexReq
	if err := ctx.ReadBody(&req); err != nil {
		responseWithError(ctx, pkt.Status_InvalidPacketBody, err)
		return
	}
	resp, err := h.msgService.GetMessageIndex(ctx.Session().GetApp(), &rpc.GetOfflineMessageIndexReq{
		Account:   ctx.Session().GetAccount(),
		MessageId: req.GetMessageId(),
	})
	if err != nil {
		responseWithError(ctx, pkt.Status_SystemException, err)
		return
	}
	var list = make([]*pkt.MessageIndex, len(resp.List))
	for i, val := range resp.List {
		list[i] = &pkt.MessageIndex{
			MessageId: val.MessageId,
			Direction: val.Direction,
			SendTime:  val.SendTime,
			AccountB:  val.AccountB,
			Group:     val.Group,
		}
	}
	err = ctx.Resp(pkt.Status_Success, &pkt.MessageIndexResp{
		Indexes: list,
	})
	if err != nil {
		responseWithError(ctx, pkt.Status_SystemException, err)
		return
	}
}

func (h *OfflineHandler) DoSyncContent(ctx cim.Context) {
	var req pkt.MessageContentReq
	if err := ctx.ReadBody(&req); err != nil {
		responseWithError(ctx, pkt.Status_InvalidPacketBody, err)
		return
	}
	if len(req.MessageIds) == 0 {
		responseWithError(ctx, pkt.Status_InvalidPacketBody, errors.New("empty MessageIds"))
		return
	}
	resp, err := h.msgService.GetMessageContent(ctx.Session().GetApp(), &rpc.GetOfflineMessageContentReq{
		MessageIds: req.MessageIds,
	})
	if err != nil {
		responseWithError(ctx, pkt.Status_SystemException, err)
		return
	}
	var list = make([]*pkt.MessageContent, len(resp.List))
	for i, val := range resp.List {
		list[i] = &pkt.MessageContent{
			MessageId: val.Id,
			Type:      val.Type,
			Body:      val.Body,
			Extra:     val.Extra,
		}
	}
	err = ctx.Resp(pkt.Status_Success, &pkt.MessageContentResp{
		Contents: list,
	})
	if err != nil {
		responseWithError(ctx, pkt.Status_SystemException, err)
		return
	}
}
