package echo

import (
	"bytes"
	cim "cirno-im"
	"context"
	"time"

	"cirno-im/demo/dialer"
	"cirno-im/logger"
	"cirno-im/websocket"
	"cirno-im/wire"
	"cirno-im/wire/pkt"
	"github.com/spf13/cobra"
)

// StartOptions StartOptions
type StartOptions struct {
}

// NewCmd NewCmd
func NewCmd(ctx context.Context) *cobra.Command {
	opts := &StartOptions{}

	cmd := &cobra.Command{
		Use:   "echo",
		Short: "Start echo client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(ctx, opts)
		},
	}

	return cmd
}

func run(ctx context.Context, opts *StartOptions) error {
	cli := websocket.NewClient("test1", "echo", nil, websocket.ClientOptions{
		Heartbeat: time.Second * 30,
		ReadWait:  time.Minute * 3,
		WriteWait: time.Second * 10,
	})

	cli.SetDialer(&dialer.ClientDialer{})

	err := cli.Connect("ws://localhost:8000")
	if err != nil {
		return err
	}
	count := 5

	go func() {
		// step3: 发送消息然后退出
		for i := 0; i < count; i++ {
			p := pkt.New(wire.CommandChatUserTalk, pkt.WithDest("test1"))
			p.WriteBody(&pkt.MessageRequest{
				Type: 1,
				Body: "hello world",
			})
			err := cli.Send(pkt.Marshal(p))
			if err != nil {
				logger.Error(err)
				return
			}
			time.Sleep(time.Second)
		}
	}()

	// step4: 接收Ack消息
	recv := 0
	for {
		frame, err := cli.Read()
		if err != nil {
			logger.Info(err)
			break
		}
		if frame.GetOpCode() != cim.OpBinary {
			continue
		}
		recv++

		p, err := pkt.MustReadLogicPkt(bytes.NewBuffer(frame.GetPayload()))
		if err != nil {
			logger.Info(err)
			break
		}
		if p.Status != pkt.Status_Success {
			var errResp pkt.ErrorResponse
			_ = p.ReadBody(&errResp)

			logger.Warnf("%s error:%s", cli.ServiceID(), errResp.Message)
		} else {
			if p.Flag == pkt.Flag_Response {
				var ack = new(pkt.MessageResponse)
				_ = p.ReadBody(ack)

				logger.Warnf("%s receive Ack [%d]", cli.ServiceID(), ack.GetMessageID())
			} else if p.Flag == pkt.Flag_Push {
				var push = new(pkt.MessagePush)
				_ = p.ReadBody(push)

				logger.Warnf("%s receive message [%d] %s", cli.ServiceID(), push.GetMessageID(), push.Body)
			}

		}

		if recv == count*2 { // 接收完消息
			break
		}
	}
	cli.Close()

	return nil
}
