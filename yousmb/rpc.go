package yousmb

import (
	"context"
	"errors"
	"github.com/project-xpolaris/youplustoolkit/yousmb/rpc"
	"github.com/projectxpolaris/youplus/config"
	"time"
)

var DefaultYouSMBRPCClient *rpc.YouSMBRPCClient

func InitYouSMCRPCClient() error {
	client := rpc.NewYouSMBRPCClient(config.Config.YouSMBRPC)
	client.KeepAlive = true
	client.MaxRetry = 99999999
	err := client.Connect(context.Background())
	if err != nil {
		return err
	}
	DefaultYouSMBRPCClient = client
	infoReply, err := client.Client.GetInfo(GetRPCTimeoutContext(), &rpc.Empty{})
	if err != nil {
		return err
	}
	if !infoReply.GetSuccess() {
		return errors.New("get yousmb rpc info failed")
	}
	return nil
}

func GetRPCTimeoutContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	return ctx
}
