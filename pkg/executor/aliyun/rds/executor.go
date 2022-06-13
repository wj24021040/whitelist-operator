package rds

import (
	"fmt"
	"strings"
	"sync"
	. "whitelist-operator/pkg/errors"

	. "whitelist-operator/pkg/executor"
	. "whitelist-operator/pkg/executor/aliyun"

	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	rds "github.com/alibabacloud-go/rds-20140815/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

const (
	secretPath = "/aliyunSecret"
	akFile     = "accessKeyId"
	skFile     = "accessKeySecret"

	DBInstanceIPArrayName = "DBInstanceIPArrayName"
	SecurityIPType        = "SecurityIPType"
	WhitelistNetworkType  = "WhitelistNetworkType"

	ALI = "aliyun"
	RDS = "rds"
)

var (
	done uint32
	l    sync.Mutex
)

type aliRdsEx struct {
	*rds.Client
}

func init() {
	//if configmap contain the ali ak/sk.  make the client
	fmt.Println("rds init")
	ak, sk, err := ReadASK(secretPath, akFile, skFile)
	if err != nil {
		return
	}

	fmt.Println(ak, sk)

	cli, err := CreateClient(&ak, &sk)
	if err != nil {
		return
	}
	fmt.Println("aliyun cleint register : ", ak, sk)
	ac := &aliRdsEx{
		Client: cli,
	}

	Register(ALI, RDS, ac)
}

func CreateClient(ak, sk *string) (*rds.Client, error) {
	if *ak == "" || *sk == "" {
		return nil, fmt.Errorf("ak/sk in empty")
	}
	var cli *rds.Client
	var err error

	config := &openapi.Config{
		// 您的AccessKey ID
		AccessKeyId: ak,
		// 您的AccessKey Secret
		AccessKeySecret: sk,
	}
	config.Endpoint = tea.String("rds.aliyuncs.com")
	// 超时设置，该产品部分接口调用比较慢，请您适当调整超时时间。
	config.ReadTimeout = tea.Int(50000)
	config.ConnectTimeout = tea.Int(50000)
	cli, err = rds.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("%w: %s (%s)\n", CREATFAIL, "the aliyun/rds cleint", err.Error())
	}

	return cli, nil
}

func (a *aliRdsEx) AddExec(param map[string]string) error {
	modifySecurityIpsRequest := a.makeReq(param)
	if modifySecurityIpsRequest == nil {
		return fmt.Errorf("the SecurityIps is empty")
	}

	modifySecurityIpsRequest.ModifyMode = tea.String("Append")
	_, err := a.ModifySecurityIps(modifySecurityIpsRequest)
	if err != nil {
		if strings.Contains(err.Error(), RdsIpDuplicateErr.Error()) {
			describeDBInstanceIPArrayListRequest := &rds.DescribeDBInstanceIPArrayListRequest{
				DBInstanceId: tea.String(*(modifySecurityIpsRequest.DBInstanceId)),
			}
			res, err := a.DescribeDBInstanceIPArrayList(describeDBInstanceIPArrayListRequest)
			if err != nil {
				return err
			}
			IPArrayName := "default"
			if modifySecurityIpsRequest.DBInstanceIPArrayName != nil {
				IPArrayName = *(modifySecurityIpsRequest.DBInstanceIPArrayName)
			}

			var existIplist string
			for _, i := range res.Body.Items.DBInstanceIPArray {
				if *(i.DBInstanceIPArrayName) == IPArrayName {
					existIplist = *(i.SecurityIPList)
				}
			}
			if existIplist == "" {
				return RdsIpDuplicateErr
			}
			newAddlist := []string{}
			ipsplice := strings.Split(*(modifySecurityIpsRequest.SecurityIps), ",")
			for _, v := range ipsplice {
				if strings.Contains(existIplist, v) {
					continue
				}
				newAddlist = append(newAddlist, v)
			}
			modifySecurityIpsRequest.SecurityIps = tea.String(strings.Join(newAddlist, ","))
			_, err = a.ModifySecurityIps(modifySecurityIpsRequest)
			if err != nil {
				return err
			}
			return nil
		}

		return err
	}
	return nil
}

func (a *aliRdsEx) DeleteExec(param map[string]string) error {
	modifySecurityIpsRequest := a.makeReq(param)
	if modifySecurityIpsRequest == nil {
		return fmt.Errorf("the SecurityIps is empty")
	}

	modifySecurityIpsRequest.ModifyMode = tea.String("Delete")
	_, err := a.ModifySecurityIps(modifySecurityIpsRequest)
	if err != nil {
		if strings.Contains(err.Error(), RdsIpNotFound.Error()) {
			return nil
		}
		return err
	}
	return nil
}

func (a *aliRdsEx) makeReq(param map[string]string) *rds.ModifySecurityIpsRequest {
	modifySecurityIpsRequest := &rds.ModifySecurityIpsRequest{}
	if _, ok := param[REGIP]; !ok {
		return nil
	}
	for k, v := range param {
		switch strings.ToLower(k) {
		case strings.ToLower(DBInstanceIPArrayName):
			modifySecurityIpsRequest.DBInstanceIPArrayName = tea.String(v)
		case strings.ToLower(SERVICEID):
			modifySecurityIpsRequest.DBInstanceId = tea.String(v)
		case strings.ToLower(REGIP):
			modifySecurityIpsRequest.SecurityIps = tea.String(v)
		case strings.ToLower(SecurityIPType):
			modifySecurityIpsRequest.SecurityIPType = tea.String(v)
		case strings.ToLower(WhitelistNetworkType):
			modifySecurityIpsRequest.WhitelistNetworkType = tea.String(v)
		case strings.ToLower(SecurityIPType):
			modifySecurityIpsRequest.SecurityIPType = tea.String(v)
		}
	}
	return modifySecurityIpsRequest
}

func (a *aliRdsEx) Default(param map[string]string) {
	if _, ok := param[DBInstanceIPArrayName]; !ok {
		param[DBInstanceIPArrayName] = "default"
	}
}

func (a *aliRdsEx) Valid(param map[string]string) error {
	for k, _ := range param {
		switch k {
		case DBInstanceIPArrayName, SecurityIPType, WhitelistNetworkType:
		default:
			return fmt.Errorf("%w for the param: %s", ParamNotSupport, k)
		}
	}
	return nil
}

func (a *aliRdsEx) IsSame(src, dst map[string]string) bool {
	if src[DBInstanceIPArrayName] == dst[DBInstanceIPArrayName] && src[SERVICEID] == dst[SERVICEID] {
		return true
	}
	return false
}
