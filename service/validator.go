package service

import (
	"bufio"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"os"
	"server/common/model"
	"server/common/utils"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func NewWatcher(file string) (chan time.Time, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	ch := make(chan time.Time)
	err = watcher.Add(file)
	if err != nil {
		return nil, err
	}
	go func() {
		defer watcher.Close()
		for {
			select {
			case e := <-watcher.Events:
				if now := time.Now(); e.Op == fsnotify.Write {
					ch <- now
				}
			case err := <-watcher.Errors:
				log.Println("watch file err: ", err)
			}
		}
	}()
	return ch, nil
}

func initValidator(db *gorm.DB) error {
	addLogFile := utils.ExpandPath("~/ops/log_add_node.log")
	addCh, err := NewWatcher(addLogFile)
	if err != nil {
		return err
	}
	go func() {
		lastLine, lastSize := int64(0), int64(0)
		for {
			stat, err := os.Stat(addLogFile)
			if err == nil {
				if newSize := stat.Size(); lastSize != newSize {
					if lastSize > newSize {
						lastLine = 0
					}
					lastSize = newSize
					file, err := os.Open(addLogFile)
					if err == nil {
						scanner := bufio.NewScanner(file)
						for i := int64(0); scanner.Scan(); i++ {
							if i >= lastLine {
								lastLine++
								split := strings.Split(scanner.Text(), " ")
								location := model.Location{
									Address: strings.ToLower(split[1]),
									IP:      split[2],
								}
								if location.IP != "127.0.0.1" {
									_, location.Latitude, location.Longitude = utils.IP2Location(location.IP)
									db.Clauses(clause.OnConflict{
										DoUpdates: clause.AssignmentColumns([]string{"ip", "latitude", "longitude"}),
									}).Create(&location)
								}
							}
						}
					}
				}
			}
			<-addCh
		}
	}()
	return nil
}

type Msg struct {
	Timestamp int64
	From      string
	To        string
}

var lastMsg []Msg

func GetLastMsg() []Msg {
	return lastMsg
}
