package service

import (
	"cirno-im/logger"
	"cirno-im/wire/rpc"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"google.golang.org/protobuf/proto"
	"time"
)

type Group interface {
	Create(app string, req *rpc.CreateGroupReq) (*rpc.CreateGroupResp, error)
	Members(app string, req *rpc.GroupMembersReq) (*rpc.GroupMembersResp, error)
	Join(app string, req *rpc.JoinGroupReq) error
	Quit(app string, req *rpc.QuitGroupReq) error
	Detail(app string, req *rpc.GetGroupReq) (*rpc.GetGroupResp, error)
}

type GroupHttp struct {
	url string
	cli *resty.Client
	srv *resty.SRVRecord
}

func NewGroupService(url string) Group {
	cli := resty.New().SetRetryCount(3).SetTimeout(time.Second * 5)
	cli.SetHeaders(map[string]string{
		"Content-Type": "application/x-protobuf",
		"Accept":       "application/x-protobuf",
	})
	cli.SetScheme("http")
	return &GroupHttp{
		url: url,
		cli: cli,
	}
}

func NewGroupServiceWithSRV(scheme string, srv *resty.SRVRecord) Group {
	cli := resty.New().SetRetryCount(3).SetTimeout(time.Second * 5)
	cli.SetHeaders(map[string]string{
		"Content-Type": "application/x-protobuf",
		"Accept":       "application/x-protobuf",
	})
	cli.SetScheme("http")
	return &GroupHttp{
		url: "",
		cli: cli,
		srv: srv,
	}
}

func (g *GroupHttp) Create(app string, req *rpc.CreateGroupReq) (*rpc.CreateGroupResp, error) {
	path := fmt.Sprintf("%s/api/%s/group", g.url, app)
	body, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	response, err := g.Request().SetBody(body).Post(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, errors.New(response.Status())
	}
	var resp rpc.CreateGroupResp
	err = proto.Unmarshal(response.Body(), &resp)
	if err != nil {
		return nil, err
	}
	logger.Debugf("GroupHttp,Create resp:%v", &resp)
	return &resp, err
}

func (g *GroupHttp) Members(app string, req *rpc.GroupMembersReq) (*rpc.GroupMembersResp, error) {
	path := fmt.Sprintf("%s/api/%s/group/member/%s", g.url, app, req.GroupId)
	response, err := g.Request().Get(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, errors.New(response.Status())
	}
	var resp rpc.GroupMembersResp
	err = proto.Unmarshal(response.Body(), &resp)
	if err != nil {
		return nil, err
	}
	logger.Debugf("GroupHttp,Members resp:%v", &resp)
	return &resp, err
}

func (g *GroupHttp) Join(app string, req *rpc.JoinGroupReq) error {
	path := fmt.Sprintf("%s/api/%s/group/member", g.url, app)
	body, _ := proto.Marshal(req)
	response, err := g.Request().SetBody(body).Post(path)
	if err != nil {
		return err
	}
	if response.StatusCode() != 200 {
		return fmt.Errorf("GroupHttp.Join response.StatusCode() = %d, want 200", response.StatusCode())
	}
	return nil
}

func (g *GroupHttp) Quit(app string, req *rpc.QuitGroupReq) error {
	path := fmt.Sprintf("%s/api/%s/group/member", g.url, app)
	body, _ := proto.Marshal(req)
	response, err := g.Request().SetBody(body).Delete(path)
	if err != nil {
		return err
	}
	if response.StatusCode() != 200 {
		return fmt.Errorf("GroupHttp.Quit response.StatusCode() = %d, want 200", response.StatusCode())
	}
	return nil
}

func (g *GroupHttp) Detail(app string, req *rpc.GetGroupReq) (*rpc.GetGroupResp, error) {
	path := fmt.Sprintf("%s/api/%s/group/%s", g.url, app, req.GroupId)
	response, err := g.Request().Get(path)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("GroupHttp.Detail response.StatusCode() = %d, want 200", response.StatusCode())
	}
	var resp rpc.GetGroupResp
	_ = proto.Unmarshal(response.Body(), &resp)
	logger.Debugf("GroupHttp.Detail resp: %v", &resp)
	return &resp, nil
}

func (g *GroupHttp) Request() *resty.Request {
	if g.srv == nil {
		return g.cli.R()
	}
	return g.cli.R().SetSRV(g.srv)
}
