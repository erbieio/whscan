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
		updateLastMsg(msgLogFile)
		for {
			select {
			case event := <-w.Events:
				if event.Name == addLogFile {
					lastLine, lastSize = updateLocation(db, addLogFile, lastLine, lastSize)
				} else if event.Name == onlineFile {
					updateOnline(db, onlineFile)
				} else if event.Name == msgLogFile {
					updateLastMsg(msgLogFile)
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
				addrs := map[string]bool{}
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
								addrs[strings.ToLower(splits[1])] = false
							}
						}
					}
				}
				var proxies []string
				db.Model(&model.Validator{}).Pluck("proxy", &proxies)
				for _, proxy := range proxies {
					if _, ok := addrs[proxy]; ok {
						addrs[proxy] = true
					}
				}
				count := 0
				for addr, isValidator := range addrs {
					if isValidator {
						count++
						log.Println("ip handler: proxy", addr, "is validator")
					} else {
						log.Println("ip handler: proxy", addr, "not validator")
					}
				}
				log.Println("ip handler:", "update", count, "of number", len(addrs))
			}
		}
	}
	return lastLine, lastSize
}

func updateOnline(db *gorm.DB, fileName string) {
	if file, err := os.Open(fileName); err == nil {
		if data, err := io.ReadAll(file); err == nil {
			peers := map[string]struct{}{}
			err := json.Unmarshal(data, &peers)
			if err == nil {
				db.Model(&model.Validator{}).Where("true").Update("online", false)
				for addr := range peers {
					addr = strings.ToLower(addr)
					changed := db.Model(&model.Validator{}).Where("`address`=?", addr).Update("online", true).RowsAffected
					if changed == 1 {
						log.Println("online handler: validator", addr, "ok")
					} else {
						log.Println("online handler: validator", addr, "not")
					}
				}
				db.Model(&model.Validator{}).Where("online=true").Count(&stats.TotalValidatorOnline)
				log.Println("online handler:", "update", stats.TotalValidatorOnline, "of number", len(peers))
			}
		}
	}
}

type Msg struct {
	From string `json:"from"`
	To   string `json:"to"`
}

var lastMsg []*Msg

func updateLastMsg(fileName string) {
	if file, err := os.Open(fileName); err == nil {
		lastMsg = lastMsg[:0]
		for scanner := bufio.NewScanner(file); scanner.Scan(); {
			splits := strings.Split(scanner.Text(), " ")
			if len(splits) == 4 {
				lastMsg = append(lastMsg, &Msg{
					From: strings.ToLower(splits[2]),
					To:   strings.ToLower(splits[3]),
				})
			}
		}
	}
}

func GetLastMsg() []*Msg {
	return lastMsg
}
