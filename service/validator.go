package service

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"server/common/model"
	"server/common/utils"
)

func initValidator(db *gorm.DB) error {
	addLogFile := utils.ExpandPath("~/ops/log_add_node.log")
	onlineFile := utils.ExpandPath("~/ops/peer.json")
	msgLogFile := utils.ExpandPath("~/ops/log_com.log")
	w, err := utils.NewWatcher([]string{addLogFile, onlineFile, msgLogFile})
	if err != nil {
		return err
	}
	go func() {
		lastLine, lastSize := updateLocation(db, addLogFile, 0, 0)
		updateOnline(db, onlineFile)
		updateLastMsg(db, msgLogFile)
		for {
			select {
			case event := <-w.Events:
				if event.Name == addLogFile {
					lastLine, lastSize = updateLocation(db, addLogFile, lastLine, lastSize)
				} else if event.Name == onlineFile {
					updateOnline(db, onlineFile)
				} else if event.Name == msgLogFile {
					updateLastMsg(db, msgLogFile)
				}
			case err := <-w.Errors:
				log.Printf("file watcher error: %v\n", err)
			}
		}
	}()
	return nil
}

func updateLocation(db *gorm.DB, fileName string, lastLine, lastSize int64) (int64, int64) {
	if stat, err := os.Stat(fileName); err == nil {
		if stat.Size() != lastSize {
			if stat.Size() < lastSize {
				lastLine = 0
			}
			lastSize = stat.Size()
			if file, err := os.Open(fileName); err == nil {
				scanner := bufio.NewScanner(file)
				for i := int64(0); scanner.Scan(); i++ {
					if i >= lastLine {
						lastLine++
						splits := strings.Split(scanner.Text(), " ")
						if len(splits) == 3 {
							if ip := splits[2]; ip != "127.0.0.1" && ip != "0.0.0.0" {
								_, latitude, longitude := utils.IP2Location(ip)
								db.Clauses(clause.OnConflict{
									DoUpdates: clause.AssignmentColumns([]string{"ip", "latitude", "longitude"}),
								}).Create(&model.Location{
									Address:   strings.ToLower(splits[1]),
									IP:        ip,
									Latitude:  latitude,
									Longitude: longitude,
								})
							}
						}
					}
				}
			}
		}
	}
	return lastLine, lastSize
}

func updateOnline(db *gorm.DB, fileName string) {
	if file, err := os.Open(fileName); err == nil {
		if data, err := io.ReadAll(file); err == nil {
			peers := map[string]struct{}{}
			err = json.Unmarshal(data, &peers)
			if err == nil {
				db.Model(&model.Validator{}).Where("true").Update("online", false)
				for addr := range peers {
					addr = strings.ToLower(addr)
					db.Model(&model.Validator{}).Where("`address`=?", addr).Update("online", true)
				}
				db.Model(&model.Validator{}).Where("`online`=true").Count(&stats.TotalValidatorOnline)
			}
		}
	}
}

type Msg struct {
	From string `json:"from"`
	To   string `json:"to"`
}

var lastMsg []*Msg

func updateLastMsg(db *gorm.DB, fileName string) {
	if file, err := os.Open(fileName); err == nil {
		lastMsg = lastMsg[:0]
		data, proxies := []string(nil), map[string]bool{}
		db.Model(&model.Validator{}).Select("proxy").Scan(&data)
		for _, proxy := range data {
			proxies[proxy] = true
		}
		for scanner := bufio.NewScanner(file); scanner.Scan(); {
			splits := strings.Split(scanner.Text(), " ")
			if len(splits) == 4 {
				from, to := strings.ToLower(splits[2]), strings.ToLower(splits[3])
				if proxies[from] && proxies[to] {
					lastMsg = append(lastMsg, &Msg{
						From: from,
						To:   to,
					})
				}
			}
		}
	}
}

func GetLastMsg() []*Msg {
	return lastMsg
}
