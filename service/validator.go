package service

import (
	"bufio"
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
	msgLogFile := utils.ExpandPath("~/ops/log_com.log")
	os.MkdirAll(utils.ExpandPath("~/ops"), os.ModePerm)
	os.WriteFile(addLogFile, nil, os.ModePerm)
	os.WriteFile(msgLogFile, nil, os.ModePerm)
	w, err := utils.NewWatcher([]string{addLogFile, msgLogFile})
	if err != nil {
		return err
	}
	go func() {
		lastLine, lastSize := updateLocation(db, addLogFile, 0, 0)
		updateLastMsg(db, msgLogFile)
		for {
			select {
			case event := <-w.Events:
				if event.Name == addLogFile {
					lastLine, lastSize = updateLocation(db, addLogFile, lastLine, lastSize)
				} else if event.Name == msgLogFile {
					updateLastMsg(db, msgLogFile)
				}
			case err := <-w.Errors:
				log.Printf("validator,file watcher error: %v\n", err)
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

type Msg struct {
	From string `json:"from"`
	To   string `json:"to"`
}

var lastMsg []*Msg

func updateLastMsg(db *gorm.DB, fileName string) {
	if file, err := os.Open(fileName); err == nil {
		lastMsg = lastMsg[:0]
		data, proxies, exist := []string(nil), map[string]bool{}, map[string]bool{}
		db.Model(&model.Validator{}).Where("amount!='0'").Select("proxy").Scan(&data)
		for _, proxy := range data {
			proxies[proxy] = true
		}
		for scanner := bufio.NewScanner(file); scanner.Scan(); {
			splits := strings.Split(scanner.Text(), " ")
			if len(splits) == 4 {
				from, to := strings.ToLower(splits[2]), strings.ToLower(splits[3])
				if proxies[from] && proxies[to] {
					if !exist[from+to] {
						lastMsg = append(lastMsg, &Msg{
							From: from,
							To:   to,
						})
						exist[from+to] = true
					}
				}
			}
		}
	}
}

// ValidatorsRes validator paging return parameters
type ValidatorsRes struct {
	Total int64             `json:"total"` //The total number of validators
	Data  []model.Validator `json:"data"`  //validator list
}

func FetchValidator(page, size int, order string) (res ValidatorsRes, err error) {
	db := DB.Where("amount!='0'")
	if order != "" {
		db = db.Order(order)
	}
	err = db.Offset((page - 1) * size).Limit(size).Find(&res.Data).Error
	res.Total = stats.TotalValidator
	return
}

type LocationRes struct {
	Address   string  `json:"address"`   //account address
	Proxy     string  `json:"proxy"`     //proxy address
	Latitude  float64 `json:"latitude"`  //latitude
	Longitude float64 `json:"longitude"` //longitude
}

func FetchLocations() (res []*LocationRes, err error) {
	err = DB.Model(&model.Validator{}).Joins("LEFT JOIN `locations` ON `validators`.`proxy`=`locations`.`address`").
		Where("`amount`!='0'").Select("`validators`.`address`,`proxy`,`latitude`,`longitude`").Scan(&res).Error
	return
}

func GetLastMsg() []*Msg {
	return lastMsg
}
