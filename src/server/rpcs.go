package main

import (
	"context"
	"slices"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/SoumyadipPayra/NightsWatch/src/db/model"
	"github.com/SoumyadipPayra/NightsWatch/src/validate"
	nwPB "github.com/SoumyadipPayra/NightsWatchProtobufs/gogenproto/nightswatch"
	"google.golang.org/grpc/metadata"
)

func (s *Server) Register(ctx context.Context, req *nwPB.RegisterRequest) (*emptypb.Empty, error) {
	if err := validate.RegisterRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error : %s", err.Error())
	}
	user := &model.User{
		UserName: req.Name,
		Password: req.Password,
		Email:    req.Email,
	}
	err := s.queryEngine.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) Login(ctx context.Context, req *nwPB.LoginRequest) (*emptypb.Empty, error) {
	if err := validate.LoginRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error : %s", err.Error())
	}
	user, err := s.queryEngine.GetUser(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	if user.Password != req.Password {
		return nil, status.Errorf(codes.Unauthenticated, "invalid username or password")
	}
	err = s.queryEngine.UpdateUserLoginTime(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error : %s", err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) SendDeviceData(ctx context.Context, req *nwPB.DeviceDataRequest) (*emptypb.Empty, error) {
	if err := validate.SendDeviceDataRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error : %s", err.Error())
	}

	// extract user id from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "expects metadata in the context")
	}
	userNames := md.Get("user_name")
	if len(userNames) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "user name not found in metadata")
	}

	userName := userNames[0]
	user, err := s.queryEngine.GetUser(ctx, userName)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	deviceData := &model.DeviceData{
		UserID: user.ID,
		InstalledApps: func(apps []*nwPB.App) []*model.App {
			appModels := make([]*model.App, len(apps))
			for i, app := range apps {
				appModels[i] = &model.App{
					Name:    app.Name,
					Version: app.Version,
				}
			}
			return appModels
		}(req.InstalledApps.GetApps()),
		OSQueryVersion: req.OsqueryVersion.GetVersion(),
		OSVersion:      req.OsVersion.GetVersion(),
		Timestamp:      time.Now(),
	}
	err = s.queryEngine.AddDeviceData(ctx, deviceData)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error : %s", err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetLatestData(ctx context.Context, req *nwPB.GetLatestDataRequest) (*nwPB.GetLatestDataResponse, error) {
	if err := validate.GetLatestDataRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error : %s", err.Error())
	}
	deviceData, err := s.queryEngine.GetLatestDeviceData(ctx, req.GetUserName())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error : %s", err.Error())
	}
	resp := &nwPB.GetLatestDataResponse{}
	if slices.Contains(req.GetDataRequestTypes(), nwPB.DeviceDataType_INSTALLED_APPS) {
		apps := make([]*nwPB.App, len(deviceData.InstalledApps))
		for i, app := range deviceData.InstalledApps {
			apps[i] = &nwPB.App{
				Name:    app.Name,
				Version: app.Version,
			}
		}
		resp.InstalledApps = &nwPB.InstalledApps{Apps: apps}
	}
	if slices.Contains(req.GetDataRequestTypes(), nwPB.DeviceDataType_OSQUERY_VERSION) {
		resp.OsqueryVersion = &nwPB.OSQueryVersion{Version: deviceData.OSQueryVersion}
	}
	if slices.Contains(req.GetDataRequestTypes(), nwPB.DeviceDataType_OS_VERSION) {
		resp.OsVersion = &nwPB.OSVersion{Version: deviceData.OSVersion}
	}
	return resp, nil
}
