package main

/* vim: set ts=2 sw=2 sts=2 ff=unix ft=go noet: */

import (
	"time"
	//"reflect"
	"strconv"
	"sync"

	"github.com/montanaflynn/stats"
	"github.com/skillcoder/homer/database"
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

func getNearestTimestamp(timestamp int64, secPeriod uint16) (ts int64) {
	ts = timestamp
	for {
		if ts%int64(secPeriod) == 0 {
			return ts
		}
		ts--
	}
}

func dbSaveMetrics(ms uint32) {
	now := time.Now()
	var secPeriod = uint16(ms / 1000)
	var timestamp = now.Unix()

	timestamp = getNearestTimestamp(timestamp, secPeriod)
	var from = timestamp - int64(secPeriod)
	rows := dbHomer.PopMetricsFrom(from, timestamp)

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

func dbProcessEvents() {
	// TODO
	// get current time in HH:mm
	// create consumption speed stats
	// save to db

}

func getNextMinute() (nextMinute int64) {
	now := time.Now()
	ts := now.Unix()
	prev := ts - (ts % 60)
	nextMinute = prev + 60
	return nextMinute
}

// resend accomulated data to clickhouse in one row per sec
func dbLoop(ms uint32) {
	dbWg.Add(1)
	defer dbWg.Done()
	ticker := time.NewTicker(time.Millisecond * time.Duration(ms))
	tickerSec := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	defer tickerSec.Stop()
	itemChan := dbQueue.GetChan()
	var nextMinute = getNextMinute()
	for {
		select {
		case item := <-itemChan:
			dbStore(item)
		case <-ticker.C:
			log.Debug("Make clickhouse row")
			dbSaveMetrics(ms)
		case <-tickerSec.C:
			if time.Now().Unix() >= nextMinute {
				nextMinute = getNextMinute()
				log.Debug("Make counter row")
				dbProcessEvents()
			}
		case <-dbShutdownChan:
			log.Debug("DB shuting down")
			// TODO make it in gorutines/parallel
			dbSaveMetrics(ms)
			dbProcessEvents()
			return
		}
	}
}

func dbShutdown() {
	dbShutdownChan <- true
	dbWg.Wait()
	log.Info("DB shutdowned")
}
