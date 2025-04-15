package main

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/SoumyadipPayra/NightsWatch/src/db/model"
	"github.com/SoumyadipPayra/NightsWatch/src/jwts"
	"github.com/SoumyadipPayra/NightsWatch/src/validate"
	nwPB "github.com/SoumyadipPayra/NightsWatchProtobufs/gogenproto/nightswatch"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func (s *Server) Register(ctx context.Context, req *nwPB.RegisterRequest) (*emptypb.Empty, error) {
	logger := zap.NewExample().Sugar()
	logger.Info("register request", zap.Any("request", req))
	if err := validate.RegisterRequest(req); err != nil {
		logger.Error("error validating register request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "error : %s", err.Error())
	}
	user := &model.User{
		UserName: req.GetName(),
		Password: req.GetPassword(),
	}
	err := s.queryEngine.CreateUser(ctx, user)
	if err != nil {
		logger.Error("error creating user", zap.Error(err))
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) Login(ctx context.Context, req *nwPB.LoginRequest) (*nwPB.LoginResponse, error) {
	logger := zap.NewExample().Sugar()
	logger.Info("login request", zap.Any("request", req))
	if err := validate.LoginRequest(req); err != nil {
		logger.Error("error validating login request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "error : %s", err.Error())
	}
	user, err := s.queryEngine.GetUser(ctx, req.Name)
	if err != nil {
		logger.Error("error getting user", zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	if user.Password != req.Password {
		logger.Error("invalid username or password")
		return nil, status.Errorf(codes.Unauthenticated, "invalid username or password")
	}
	err = s.queryEngine.UpdateUserLoginTime(ctx, req.GetName())
	if err != nil {
		logger.Error("error updating user login time", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "error : %s", err.Error())
	}

	tokenString, err := jwts.GenerateToken(req.GetName())
	if err != nil {
		logger.Error("error generating token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "error : %s", err.Error())
	}

	return &nwPB.LoginResponse{Token: tokenString}, nil
}

func (s *Server) SendDeviceData(ctx context.Context, req *nwPB.DeviceDataRequest) (*emptypb.Empty, error) {
	logger := zap.NewExample().Sugar()
	logger.Info("send device data request", zap.Any("request", req))
	if err := validate.SendDeviceDataRequest(req); err != nil {
		logger.Error("error validating send device data request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "error : %s", err.Error())
	}

	// extract user id from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Error("expects metadata in the context")
		return nil, status.Errorf(codes.InvalidArgument, "expects metadata in the context")
	}

	tokens := md.Get("jwt_token")
	if len(tokens) == 0 {
		logger.Error("jwt token not found")
		return nil, status.Errorf(codes.Unauthenticated, "jwt token not found")
	}

	userName, err := jwts.ValidateToken(tokens[0])
	if err != nil {
		logger.Error("invalid token", zap.Error(err))
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}

	user, err := s.queryEngine.GetUser(ctx, userName)
	if err != nil {
		logger.Error("error getting user", zap.Error(err))
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
		logger.Error("error adding device data", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "error : %s", err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetLatestData(ctx context.Context, req *nwPB.GetLatestDataRequest) (*nwPB.GetLatestDataResponse, error) {
	logger := zap.NewExample().Sugar()
	logger.Info("get latest data request", zap.Any("request", req))
	if err := validate.GetLatestDataRequest(req); err != nil {
		logger.Error("error validating get latest data request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "error : %s", err.Error())
	}
	deviceData, err := s.queryEngine.GetLatestDeviceData(ctx, req.GetUserName())
	if err != nil {
		logger.Error("error getting latest device data", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "error : %s", err.Error())
	}

	resp := &nwPB.GetLatestDataResponse{
		InstalledApps: &nwPB.InstalledApps{
			Apps: func(apps []*model.App) []*nwPB.App {
				appModels := make([]*nwPB.App, len(apps))
				for i, app := range apps {
					appModels[i] = &nwPB.App{
						Name:    app.Name,
						Version: app.Version,
					}
				}
				return appModels
			}(deviceData.InstalledApps),
		},
		OsqueryVersion: &nwPB.OSQueryVersion{Version: deviceData.OSQueryVersion},
		OsVersion:      &nwPB.OSVersion{Version: deviceData.OSVersion},
	}
	return resp, nil
}
