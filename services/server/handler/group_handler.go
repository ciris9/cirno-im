package handler

import (
	cim "cirno-im"
	"cirno-im/services/server/service"
	"cirno-im/wire/pkt"
	"cirno-im/wire/rpc"
	"errors"
)

type GroupHandler struct {
	groupService service.Group
}

func NewGroupHandler(groupService service.Group) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
	}
}

func (h *GroupHandler) DoCreate(ctx cim.Context) {
	var req pkt.GroupCreateRequest
	if err := ctx.ReadBody(&req); err != nil {
		responseWithError(ctx, pkt.Status_InvalidPacketBody, err)
		return
	}
	resp, err := h.groupService.Create(ctx.Session().GetApp(), &rpc.CreateGroupReq{
		Name:         req.GetName(),
		Avatar:       req.GetAvatar(),
		Introduction: req.GetIntroduction(),
		Owner:        req.GetOwner(),
		Members:      req.GetMembers(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	locations, err := ctx.GetLocations(req.GetMembers()...)
	if err != nil && !errors.Is(err, cim.ErrSessionNil) {
		responseWithError(ctx, pkt.Status_SystemException, err)
		return
	}

	// push to receiver
	if len(locations) > 0 {
		if err = ctx.Dispatch(&pkt.GroupCreateNotify{
			GroupId: resp.GroupId,
			Members: req.GetMembers(),
		}, locations...); err != nil {
			responseWithError(ctx, pkt.Status_SystemException, err)
			return
		}
	}

	err = ctx.Resp(pkt.Status_Success, &pkt.GroupCreateResponse{
		GroupId: resp.GroupId,
	})
	if err != nil {
		responseWithError(ctx, pkt.Status_SystemException, err)
		return
	}
}

func (h *GroupHandler) DoJoin(ctx cim.Context) {
	var req pkt.GroupJoinReq
	if err := ctx.ReadBody(&req); err != nil {
		responseWithError(ctx, pkt.Status_InvalidPacketBody, err)
		return
	}
	err := h.groupService.Join(ctx.Session().GetApp(), &rpc.JoinGroupReq{
		Account: req.Account,
		GroupId: req.GetGroupId(),
	})
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

func (h *GroupHandler) DoQuit(ctx cim.Context) {
	var req pkt.GroupQuitReq
	if err := ctx.ReadBody(&req); err != nil {
		responseWithError(ctx, pkt.Status_InvalidPacketBody, err)
		return
	}
	err := h.groupService.Quit(ctx.Session().GetApp(), &rpc.QuitGroupReq{
		Account: req.Account,
		GroupId: req.GetGroupId(),
	})
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

func (h *GroupHandler) DoDetail(ctx cim.Context) {
	var req pkt.GroupGetReq
	if err := ctx.ReadBody(&req); err != nil {
		responseWithError(ctx, pkt.Status_InvalidPacketBody, err)
		return
	}
	resp, err := h.groupService.Detail(ctx.Session().GetApp(), &rpc.GetGroupReq{
		GroupId: req.GetGroupId(),
	})
	if err != nil {
		responseWithError(ctx, pkt.Status_SystemException, err)
		return
	}
	membersResp, err := h.groupService.Members(ctx.Session().GetApp(), &rpc.GroupMembersReq{
		GroupId: req.GetGroupId(),
	})
	if err != nil {
		responseWithError(ctx, pkt.Status_SystemException, err)
		return
	}
	var members = make([]*pkt.Member, len(membersResp.GetUsers()))
	for i, m := range membersResp.GetUsers() {
		members[i] = &pkt.Member{
			Account:  m.Account,
			Alias:    m.Alias,
			JoinTime: m.JoinTime,
			Avatar:   m.Avatar,
		}
	}
	err = ctx.Resp(pkt.Status_Success, &pkt.GroupGetResp{
		Id:           resp.Id,
		Name:         resp.Name,
		Introduction: resp.Introduction,
		Avatar:       resp.Avatar,
		Owner:        resp.Owner,
		Members:      members,
	})
	if err != nil {
		responseWithError(ctx, pkt.Status_SystemException, err)
		return
	}
}
