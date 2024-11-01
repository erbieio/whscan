package service

import (
	"strings"
	"testing"
)

func TestCreateDir(t *testing.T) {
	addr := "0xFA623BCC71BE5C3aBacfe875E64ef97F91B7b110"
	addr = strings.ToLower(addr)
	path := "contractcode/" + addr[len(addr)-2:] + "/" + addr
	err := createDir(path)
	t.Log("err=", err)
}

func TestGetSourceCode(t *testing.T) {
	addr := "0xFA623BCC71BE5C3aBacfe875E64ef97F91B7b110"
	code, err := GetSourceCode(addr)
	t.Log("err=", err)
	t.Log("code : \n", code)
}
