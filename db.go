package main

/* vim: set ts=2 sw=2 sts=2 ff=unix ft=go noet: */

import (
	"time"
	//"reflect"
	"strconv"
	"sync"

	"github.com/montanaflynn/stats"
	"github.com/skillcoder/homer/database"
	"github.com/skillcoder/homer/helper"
)

var dbQueue *database.DbQueueT
var dbHomer *database.DbT
var dbShutdownChan = make(chan bool)
var dbWg sync.WaitGroup
var oldrow = make(map[string]float64)

func dbInit() {
	dbHomer = database.New(clickhouseMetricCount)
	dbQueue = database.NewQueue(128)
}

func dbInsert(fieldName string, fieldType string, valueInterface string, time int64) {
	dbQueue.Add(database.CreateItem(fieldName, fieldType, valueInterface, time))
}

func getClickTime(item database.DbItemT) (timestamp int64) {
	timestamp = item.Time
	value, err := strconv.ParseUint(item.ColVal, 10, 64)
	if err != nil {
		// handle error
		log.Errorf("Cant convert to time %s %s %s convert to int: %s", item.ColName, item.ColType, item.ColVal, err)
		return timestamp
	}

	if value > 1000000000 {
		timestamp = int64(value)
	}

	return timestamp
}

func dbGetStatHandler() database.StatT {
	return dbHomer.GetStat()
}

func dbStore(item database.DbItemT) {
	//value := reflect.ValueOf(item.ColVal)
	//valueType := value.Type()
	log.Debugf("DB [%d] %s:%s = %v", item.Time, item.ColName, item.ColType, item.ColVal)
	switch item.ColType {
	case "count":
		if counterID, ok := configCounters[item.ColName]; ok {
			dbHomer.AddClick(counterID, getClickTime(item))
		} else {
			log.Warn("Unknown counter:", item.ColName)
		}
	case "move":
		log.Warn("MOVE")
	case "led":
		log.Warn("LED")
	case "temp", "humd", "pres":
		dbHomer.AddMetric(item.Time, item)
	default:
		log.Warnf("Unknown item.ColType: %s", item.ColType)
	}
}

func dbSaveMetrics(ms int) {
	var timestamp = time.Now().Unix()
	var secPeriod = ms / 1000

	timestamp = helper.GetPrevTimestamp(timestamp, secPeriod)
	var from = timestamp - int64(secPeriod)
	rows, isHaveOlder := dbHomer.PopMetricsFrom(from, timestamp)
	if isHaveOlder {
		// FIXME need send it too (its late data)
		log.Warn("We have older metrics in database, need implement send it too")
	}

	row := make(map[string]float64)    // Result
	nowrow := make(map[string]float64) //next old row
	for key, arr := range rows {
		median, err := stats.Median(arr)
		if err == nil {
			nowrow[key] = median
			row[key] = median
		} else {
			log.Warnf("Cant get Median for %s: %v", key, arr)
		}
	}

	// if value is empty we get it from prev value
	for key, val := range oldrow {
		if _, ok := row[key]; !ok {
			row[key] = val
		}
	}

	// clear old row & copy to it nowrow
	oldrow := make(map[string]float64)
	for key, val := range nowrow {
		oldrow[key] = val
	}

	go clickhouseMetricInsert(timestamp, row)
}

func dbSaveClicks() {
	var timestamp = time.Now().Unix()
	var secPeriod = 60
	timestamp = helper.GetPrevTimestamp(timestamp, secPeriod)
	var from = timestamp - int64(secPeriod)
	rows, isHaveOlder := dbHomer.PopClicksFrom(from, timestamp)
	if isHaveOlder {
		// FIXME need send it too (its late data)
		log.Warn("We have older clicks in database, need implement send it too")
	}

	for counterID, count := range rows {
		log.Warnf("[%d] %1d => %d", time.Unix(timestamp, 0).Format("15:04:05"), counterID, count)
	}
}

// resend accomulated data to clickhouse in one row per sec
func dbLoop(ms int) {
	dbWg.Add(1)
	defer dbWg.Done()
	ticker := time.NewTicker(time.Millisecond * time.Duration(ms))
	tickerSec := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	defer tickerSec.Stop()
	itemChan := dbQueue.GetChan()
	var nextMinute = helper.GetNextMinute(time.Now().Unix())
	for {
		select {
		case item := <-itemChan:
			dbStore(item)
		case <-ticker.C:
			log.Debug("Make clickhouse row")
			dbSaveMetrics(ms)
		case <-tickerSec.C:
			var timestamp = time.Now().Unix()
			if timestamp >= nextMinute {
				nextMinute = helper.GetNextMinute(timestamp)
				log.Debug("Make counter row")
				dbSaveClicks()
			}
		//TODO create water consumption speed stats, every sec
		case <-dbShutdownChan:
			log.Debug("DB shuting down")
			// TODO make it in gorutines/parallel
			dbSaveMetrics(ms)
			dbSaveClicks()
			return
		}
	}
}

func dbShutdown() {
	dbShutdownChan <- true
	dbWg.Wait()
	log.Info("DB shutdowned")
}
