package handler

import (
	"cirno-im/services/service/database"
	"cirno-im/wire/rpc"
	"errors"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

func (h *ServiceHandler) GroupCreate(c iris.Context) {
	app := c.Params().Get("app")
	var req rpc.CreateGroupReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	groupId := h.IdGen.Next()
	var g = &database.Group{
		Model: database.Model{
			ID: groupId.Int64(),
		},
		Group:        groupId.Base36(),
		App:          app,
		Name:         req.Name,
		Owner:        req.Owner,
		Avatar:       req.Avatar,
		Introduction: req.Introduction,
	}
	members := make([]database.GroupMember, len(req.Members))
	for i, user := range req.Members {
		members[i] = database.GroupMember{
			Model: database.Model{
				ID: h.IdGen.Next().Int64(),
			},
			Account: user,
			Group:   groupId.Base36(),
		}
	}
	err := h.BaseDb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(g).Error; err != nil {
			return err
		}
		if err := tx.Create(&members).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	if _, err = c.Negotiate(&rpc.CreateGroupResp{
		GroupId: groupId.Base36(),
	}); err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

func (h *ServiceHandler) GroupJoin(c iris.Context) {
	var req rpc.JoinGroupReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	gm := &database.GroupMember{
		Model: database.Model{
			ID: h.IdGen.Next().Int64(),
		},
		Account: req.Account,
		Group:   req.GroupId,
	}
	err := h.BaseDb.Create(gm).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

func (h *ServiceHandler) GroupQuit(c iris.Context) {
	var req rpc.QuitGroupReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	gm := &database.GroupMember{
		Account: req.Account,
		Group:   req.GroupId,
	}
	err := h.BaseDb.Delete(&database.GroupMember{}, gm).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

func (h *ServiceHandler) GroupMembers(c iris.Context) {
	group := c.Params().Get("id")
	if group == "" {
		c.StopWithError(iris.StatusBadRequest, errors.New("group id can not be empty"))
		return
	}
	var members []database.GroupMember
	err := h.BaseDb.Order("Updated_At desc").Find(&members, database.GroupMember{
		Group: group,
	}).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	var users = make([]*rpc.Member, len(members))
	for i, member := range members {
		users[i] = &rpc.Member{
			Account:  member.Account,
			Avatar:   member.Account,
			JoinTime: member.CreatedAt.Unix(),
		}
	}
	if _, err = c.Negotiate(&rpc.GroupMembersResp{
		Users: users,
	}); err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

func (h *ServiceHandler) GroupGet(c iris.Context) {
	groupId := c.Params().Get("id")
	if groupId == "" {
		c.StopWithError(iris.StatusBadRequest, errors.New("group id can not be empty"))
		return
	}
	id, err := h.IdGen.ParseBase36(groupId)
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	var group database.Group
	err = h.BaseDb.First(&group, id).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	if _, err = c.Negotiate(&rpc.GetGroupResp{
		Id:           groupId,
		Name:         group.Name,
		Avatar:       group.Avatar,
		Introduction: group.Introduction,
		Owner:        group.Owner,
		CreatedAt:    group.CreatedAt.Unix(),
	}); err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}
