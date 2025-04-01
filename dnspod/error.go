package dnspod

import (
	"errors"

	terrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
)

var (
	ErrNeedSecretName = errors.New("secret name must be specified")
	ErrNeedSecretKey  = errors.New("secret key must be specified")
)

func isRecordNotFound(err error) bool {
	if e, ok := err.(*terrors.TencentCloudSDKError); ok {
		if e.Code == "ResourceNotFound.NoDataOfRecord" {
			return true
		}
	}
	return false
}
