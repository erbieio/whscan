package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const LOG_CHANNEL_SIZE = 10

var chanMsg = make(chan string, LOG_CHANNEL_SIZE)
var botToken = "1058353695:AAEDE92YObXRr2-dPxMhaEw0i0Hw5uy883Y"
var groupId = int64(-636377631)

//*******************************************************
// Public Function
//*******************************************************

//InitSendLogByBot 機器人日誌發送初始化
func init() {
	initChanRecLogMsg(botToken, groupId)
}

//SendMsg
func SendToBot(format string, args ...interface{}) {
	hostname, _ := os.Hostname()
	sendToChanMsg("HostName:"+hostname+"\n\r"+fmt.Sprintf(format, args...), chanMsg)
}

//*******************************************************
// Private Function
//*******************************************************

//initChanRecLogMsg 初始化channel接受消息
func initChanRecLogMsg(BotToken string, groupId int64) {
	go func() {
		for m := range chanMsg {
			time.Sleep(time.Millisecond * 1)
			for i := 0; i < len(m); i += 3000 {
				var s string
				if len(m) < i+3000 {
					s = m[i:]
				} else {
					s = m[i : i+3000]
				}
				sendToBotServer(s, BotToken, groupId)
			}
		}
	}()
	fmt.Printf("init send bot msg. groupId:%d botToken:%s\n", groupId, botToken)
}

//sendToChanMsg 發送到channel
func sendToChanMsg(msg string, chanmsg chan string) {
	select {
	case chanmsg <- msg:
	default:
		Errorf("sendToChanMsg overflow! msg:%v\n", msg)
	}
}

//sendToBotServer post發送到機器人服務
func sendToBotServer(msg string, token string, groupid int64) error {
	req := make(map[string]interface{})
	req["chat_id"] = groupid
	req["text"] = msg
	//fmt.Println("req: ", req)
	data, err := json.Marshal(req)
	if err != nil {
		Errorf("req to json is err: %v\n", err)
		return err
	}

	reader := bytes.NewReader(data)
	request, err := http.NewRequest("POST", "https://api.telegram.org/bot"+token+"/sendMessage", reader)
	if err != nil {
		Errorf("post telegram chat msg err: %v\n", err)
		return err
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		request.Header.Set("Content-Type", "application/json;charset=UTF-8")
		client := http.Client{}
		resp, err := client.Do(request.WithContext(ctx))
		if err != nil {
			Errorf("send message to tg error:%v\n", err)
			return err
		} else {
			data, _ := ioutil.ReadAll(resp.Body)
			Debugf("tg api resp: %s \n", data)
			resp.Body.Close()
		}
	}

	return nil
}
