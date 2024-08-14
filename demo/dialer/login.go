package dialer

import (
	cim "cirno-im"
	"cirno-im/websocket"
	"cirno-im/wire/token"
)

func Login(wsurl, account string, appSecrets ...string) (cim.Client, error) {
	cli := websocket.NewClient(account, "unittest", websocket.ClientOptions{})
	secret := token.DefaultSecret
	if len(appSecrets) > 0 {
		secret = appSecrets[0]
	}
	cli.SetDialer(&ClientDialer{
		AppSecret: secret,
	})
	err := cli.Connect(wsurl)
	if err != nil {
		return nil, err
	}
	return cli, nil
}
