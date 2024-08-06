package handler

import (
	"cirno-im/services/service/database"
	"cirno-im/wire/rpc"
	"github.com/go-redis/redis/v7"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

type ServiceHandler struct {
	BaseDb    *gorm.DB
	MessageDb *gorm.DB
	Cache     *redis.Client
	IdGen     *database.IDGenerator
}

func (h *ServiceHandler) InsertUserMessage(c iris.Context) {
	var req rpc.InsertMessageReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}
	messageId := h.IdGen.Next().Int64()
	messageContent := database.MessageContent{
		ID:       messageId,
		Type:     byte(req.Message.Type),
		Body:     req.Message.Body,
		Extra:    req.Message.Extra,
		SendTime: req.SendTime,
	}
	//扩散写
	idxs := make([]database.MessageIndex, 2)
	idxs[0] = database.MessageIndex{
		ID:        h.IdGen.Next().Int64(),
		AccountA:  req.Dest,
		AccountB:  req.Sender,
		Direction: 0,
		MessageID: messageId,
		SendTime:  req.SendTime,
	}
	idxs[1] = database.MessageIndex{
		ID:        h.IdGen.Next().Int64(),
		AccountA:  req.Sender,
		AccountB:  req.Dest,
		Direction: 1,
		MessageID: messageId,
		SendTime:  req.SendTime,
	}
	err := h.MessageDb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&idxs).Error; err != nil {
			return err
		}
		if err := tx.Create(&messageContent).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	if _, err = c.Negotiate(&rpc.InsertMessageResp{
		MessageId: messageId,
	}); err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

func (h *ServiceHandler) InsertGroupMessage(c iris.Context) {
	var req rpc.InsertMessageReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	messageId := h.IdGen.Next().Int64()
	var members []database.GroupMember
	err := h.BaseDb.Where(&database.GroupMember{Group: req.Dest}).Find(&members).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	var idxs = make([]database.MessageIndex, len(members))
	for i, member := range members {
		idxs[i] = database.MessageIndex{
			ID:        h.IdGen.Next().Int64(),
			AccountA:  member.Account,
			AccountB:  req.Sender,
			Direction: 0,
			MessageID: messageId,
			Group:     member.Group,
			SendTime:  req.SendTime,
		}
		if member.Account == req.Sender {
			idxs[i].Direction = 1
		}
	}
	messageContent := database.MessageContent{
		ID:       messageId,
		Type:     byte(req.Message.Type),
		Body:     req.Message.Body,
		Extra:    req.Message.Extra,
		SendTime: req.SendTime,
	}
	err = h.MessageDb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&idxs).Error; err != nil {
			return err
		}
		if err := tx.Create(&messageContent).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
	if _, err = c.Negotiate(&rpc.InsertMessageResp{MessageId: messageId}); err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}
