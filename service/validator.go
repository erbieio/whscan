package service

type Msg struct {
	Timestamp int64
	From      string
	To        string
}

var lastMsg []Msg

func GetLastMsg() []Msg {
	return lastMsg
}
