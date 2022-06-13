package errors

import (
	"errors"
)

var (
	PNOTEXIST = errors.New("the executor prvider is not exist")
	SNOTEXIST = errors.New("the executor service is not exist")
	CREATFAIL = errors.New("create the object failed")

	RdsIpDuplicateErr = errors.New("InvalidInstanceIp.Duplicate")
	RdsIpNotFound     = errors.New("InvalidSecurityIPs.NotFound")

	ParamNotSupport = errors.New("ParamNotSupport")
)
