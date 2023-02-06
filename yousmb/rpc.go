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
	err := client.Init()
	if err != nil {
		return err
	}
	rpcClient, _, err := client.GetClient()
	if err != nil {
		return err
	}
	DefaultYouSMBRPCClient = client
	infoReply, err := rpcClient.GetInfo(GetRPCTimeoutContext(), &rpc.Empty{})
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

func ExecWithRPCClient(fn func(client rpc.YouSMBServiceClient) error) error {
	client, conn, err := DefaultYouSMBRPCClient.GetClient()
	if err != nil {
		return err
	}
	defer conn.Close()
	return fn(client)
}
