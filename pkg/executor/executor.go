package executor

import (
	"fmt"
	. "whitelist-operator/pkg/errors"
)

const (
	ADD   = "add"
	DEL   = "del"
	CHECK = "check"
	DEF   = "default"

	REGIP     = "registerIp"
	SERVICEID = "serviceId"
)

type Executor interface {
	AddExec(param map[string]string) error
	DeleteExec(param map[string]string) error
	Valid(param map[string]string) error
	Default(param map[string]string)
	IsSame(src, dst map[string]string) bool
}

var Executors map[string]map[string]Executor = make(map[string]map[string]Executor)

func Register(provider, service string, exec Executor) {
	s, ok := Executors[provider]
	if !ok {
		s = make(map[string]Executor)
		Executors[provider] = s
	}

	s[service] = exec
}

func Exec(ops, provider, service string, param map[string]string) error {
	s, ok := Executors[provider]
	if !ok {
		return fmt.Errorf("%w, prvider: %s\n", PNOTEXIST, provider)
	}
	e, ok := s[service]
	if !ok {
		return fmt.Errorf("%w, prvider: %s, service: %s\n", SNOTEXIST, provider, service)
	}
	var err error
	switch ops {
	case ADD:
		err = e.AddExec(param)
	case DEL:
		err = e.DeleteExec(param)
	case CHECK:
		err = e.Valid(param)
	case DEF:
		e.Default(param)
	}
	return err
}

func Duplicate(provider, service string, src, dst map[string]string) bool {
	s, ok := Executors[provider]
	if !ok {
		return false
	}
	e, ok := s[service]
	if !ok {
		return false
	}
	return e.IsSame(src, dst)
}
