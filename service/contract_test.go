package service

import "testing"

func TestCreateDir(t *testing.T) {
	addr := "0xFA623BCC71BE5C3aBacfe875E64ef97F91B7b110"
	path := "contractcode/" + addr[len(addr)-2:] + "/" + addr
	err := createDir(path)
	t.Log("err=", err)
}
