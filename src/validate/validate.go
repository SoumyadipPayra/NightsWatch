package validate

import (
	nwPB "github.com/SoumyadipPayra/NightsWatchProtobufs/gogenproto/nightswatch"
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func RegisterRequest(req *nwPB.RegisterRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Name, validation.Required),
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.Password, validation.Required),
	)
}

func LoginRequest(req *nwPB.LoginRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.Name, validation.Required),
		validation.Field(&req.Password, validation.Required),
	)
}

func SendDeviceDataRequest(req *nwPB.DeviceDataRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.InstalledApps, validation.Required),
		validation.Field(&req.OsVersion, validation.Required),
		validation.Field(&req.OsqueryVersion, validation.Required),
	)
}

func GetLatestDataRequest(req *nwPB.GetLatestDataRequest) error {
	return validation.ValidateStruct(req,
		validation.Field(&req.UserName, validation.Required),
		validation.Field(&req.DataRequestTypes, validation.Required),
	)
}
