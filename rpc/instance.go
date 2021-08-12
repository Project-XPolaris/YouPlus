package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/projectxpolaris/youplus/config"
	"github.com/projectxpolaris/youplus/service"
	"github.com/projectxpolaris/youplus/utils"
	"google.golang.org/grpc"
	"log"
	"net"
	"strings"
)

var DefaultRPCServer = &RPCServer{}

const (
	ErrorCodeUnknown        = 9999
	ErrorCodeEntityNotFound = 6001
)

type RPCServer struct {
	server Server
}

func (l *RPCServer) Run() {
	lis, err := net.Listen("tcp", config.Config.RPCAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	rpcServer := grpc.NewServer()
	l.server = Server{}
	RegisterYouPlusServiceServer(rpcServer, &l.server)
	log.Printf("server listening at %v", lis.Addr())
	if err := rpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type Server struct {
	UnimplementedYouPlusServiceServer
}

func (s Server) CheckDataset(ctx context.Context, in *CheckDatasetRequest) (*CheckDatasetReply, error) {
	dataset, err := service.DefaultZFSManager.GetDatasetByPath(in.GetPath())
	if err != nil {
		return nil, err
	}
	isDataset := dataset != nil
	return &CheckDatasetReply{
		IsDataset: &isDataset,
	}, nil

}
func (s Server) GetDatasetInfo(ctx context.Context, in *GetDatasetInfoRequest) (*GetDatasetInfoReply, error) {
	dataset, err := service.DefaultZFSManager.GetDatasetByPath(*in.Dataset)
	if err != nil {
		return nil, err
	}
	if dataset == nil {
		return nil, errors.New("not a dataset")
	}
	snapshots, err := dataset.Snapshots()
	if err != nil {
		return nil, err
	}
	datasetPath, err := dataset.Path()
	if err != nil {
		return nil, err
	}
	snapshotsReplyList := make([]*Snapshot, 0)
	for _, snapshot := range snapshots {
		snapshotPath, err := snapshot.Path()
		if err != nil {
			continue
		}
		snapshotName := strings.Replace(snapshotPath, datasetPath+"@", "", 1)
		snapshotsReplyList = append(snapshotsReplyList, &Snapshot{Name: &snapshotName})
	}
	return &GetDatasetInfoReply{Snapshots: snapshotsReplyList, Path: in.Dataset}, nil
}

func (s Server) CreateDataset(ctx context.Context, in *CreateDatasetRequest) (*ActionReply, error) {
	datasetPath, err := service.DefaultZFSManager.PathToZFSPath(*in.Path)
	if err != nil {
		return nil, err
	}
	_, err = service.DefaultZFSManager.CreateDataset(datasetPath)
	if err != nil {
		return nil, err
	}
	success := true
	return &ActionReply{Success: &success}, nil
}

func (s Server) DeleteDataset(ctx context.Context, in *DeleteDatasetRequest) (*ActionReply, error) {
	datasetPath, err := service.DefaultZFSManager.PathToZFSPath(*in.Path)
	if err != nil {
		return nil, err
	}
	err = service.DefaultZFSManager.DeleteDataset(datasetPath)
	if err != nil {
		return nil, err
	}
	success := true
	return &ActionReply{Success: &success}, nil
}

func (s Server) CreateSnapshot(ctx context.Context, in *CreateSnapshotRequest) (*ActionReply, error) {
	datasetPath, err := service.DefaultZFSManager.PathToZFSPath(*in.Dataset)
	if err != nil {
		return nil, err
	}
	_, err = service.DefaultZFSManager.CreateSnapshot(datasetPath, *in.Snapshot)
	if err != nil {
		return nil, err
	}
	success := true
	return &ActionReply{Success: &success}, nil
}

func (s Server) DeleteSnapshot(ctx context.Context, in *DeleteSnapshotRequest) (*ActionReply, error) {
	datasetPath, err := service.DefaultZFSManager.PathToZFSPath(*in.Dataset)
	if err != nil {
		return nil, err
	}
	err = service.DefaultZFSManager.DeleteSnapshot(datasetPath, *in.Snapshot)
	if err != nil {
		return nil, err
	}
	success := true
	return &ActionReply{Success: &success}, nil
}

func (s Server) RollbackDataset(ctx context.Context, in *RollbackDatasetRequest) (*ActionReply, error) {
	datasetPath, err := service.DefaultZFSManager.PathToZFSPath(*in.Dataset)
	if err != nil {
		return nil, err
	}
	err = service.DefaultZFSManager.DatasetRollback(datasetPath, *in.Snapshot)
	if err != nil {
		return nil, err
	}
	success := true
	return &ActionReply{Success: &success}, nil
}

func (s Server) RegisterEntry(ctx context.Context, in *RegisterEntryRequest) (*ActionReply, error) {
	service.DefaultRegisterManager.RegisterApp(&service.Entry{
		Name:     *in.Name,
		Instance: *in.Instance,
		Version:  *in.Version,
	})
	success := true
	return &ActionReply{Success: &success}, nil
}
func (s Server) UnregisterEntry(ctx context.Context, in *UnregisterEntryRequest) (*ActionReply, error) {
	service.DefaultRegisterManager.UnregisterApp(*in.Instance)
	success := true
	return &ActionReply{Success: &success}, nil
}
func (s Server) EntryHeartbeat(ctx context.Context, in *HeartbeatRequest) (*ActionReply, error) {
	err := service.DefaultRegisterManager.Heartbeat(in.GetInstance(), in.GetState())
	if err != nil {
		code := ErrorCodeUnknown
		if err == service.EntryNotFoundError {
			code = ErrorCodeEntityNotFound
		}
		return &ActionReply{
			Success: utils.GetBoolPtr(false),
			Reason:  utils.GetStringPtr(err.Error()),
			Code:    utils.GetInt64Ptr(int64(code)),
		}, nil
	}
	success := true
	return &ActionReply{Success: &success}, nil
}

func (s Server) UpdateEntryExport(ctx context.Context, in *UpdateEntryExportRequest) (*ActionReply, error) {
	rawData := []byte(in.GetData())
	result := service.EntityExport{}
	err := json.Unmarshal(rawData, &result)
	if err != nil {
		return nil, err
	}
	err = service.DefaultRegisterManager.UpdateExport(in.GetInstance(), result)
	if err != nil {
		return nil, err
	}
	success := true
	return &ActionReply{Success: &success}, nil
}
