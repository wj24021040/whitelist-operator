package aliyun

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func ReadASK(secretPath, akFile, skFile string) (ak, sk string, err error) {
	f, err := os.Stat(secretPath)
	if err != nil || !(f.IsDir()) {
		return "", "", err
	}
	akPath := path.Join(secretPath, akFile)
	aktmp, err := ioutil.ReadFile(akPath)
	if err != nil {
		return "", "", err
	}

	skPath := path.Join(secretPath, skFile)
	sktmp, err := ioutil.ReadFile(skPath)
	if err != nil {
		return "", "", err
	}

	ak = strings.Replace(string(aktmp), "\n", "", -1)
	sk = strings.Replace(string(sktmp), "\n", "", -1)
	return
}
