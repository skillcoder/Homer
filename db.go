package main

/* vim: set ts=2 sw=2 sts=2 ff=unix ft=go noet: */

import (
	"time"
	//"reflect"
	"sort"
	"strconv"
	"sync"

	"github.com/montanaflynn/stats"
)

type dbItemT struct {
	ColName string
	ColType string
	ColVal  string
	Time    int64
}

type dbQueueT struct {
	ch chan dbItemT
}

func (q *dbQueueT) Add(item dbItemT) {
	q.ch <- item
}

func (q *dbQueueT) getChan() <-chan dbItemT {
	return q.ch
}

type dbT struct {
	m      map[int64][]dbItemT
	clicks map[uint8][]int64
}

func newDatabase() *dbT {
	return &dbT{
		m:      make(map[int64][]dbItemT),
		clicks: make(map[uint8][]int64),
	}
}

func (c *dbT) Load(key int64) ([]dbItemT, bool) {
	val, ok := c.m[key]
	return val, ok
}

func (c *dbT) AddMetric(key int64, item dbItemT) {
	column, ok := c.Load(key)
	if !ok {
		column = make([]dbItemT, 0, clickhouseMetricCount+1)
	}

	column = append(column, item)
	c.m[key] = column
}

func (c *dbT) AddClick(key uint8, time int64) {
	click, ok := c.clicks[key]
	if !ok {
		click = make([]int64, 0, 10)
	}

	click = append(click, time)
	c.clicks[key] = click
}

func (c *dbT) popMetricsFrom(from, to int64) (rows map[string][]float64) {
	keys := make([]int64, 0, len(c.m))
	for k := range c.m {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	rows = make(map[string][]float64)
	for _, key := range keys {
		if key < to {
			if key >= from {
				//log.Debugf("key [%d] %v", key, item);
				for _, item := range c.m[key] {
					rowKey := item.ColName + ":" + item.ColType
					value, err := strconv.ParseFloat(item.ColVal, 64)
					if err != nil {
						log.Errorf("[%d] %s %s %s convert to float: %s", item.Time, item.ColName, item.ColType, item.ColVal, err)
						continue
					}

					rows[rowKey] = append(rows[rowKey], value)
				}
				delete(c.m, key)
			} else {
				// FIXME need send it too (its late data)
				log.Warnf("key [%d] %v", key, c.m[key])
			}
		}
	}

	return rows
}

var dbQueue dbQueueT
var database *dbT
var dbShutdownChan = make(chan bool)
var dbWg sync.WaitGroup
var oldrow = make(map[string]float64)

func init() {
	database = newDatabase()
	dbQueue = dbQueueT{
		ch: make(chan dbItemT, 128),
	}
}

func dbInsert(fieldName string, fieldType string, valueInterface string, time int64) {
	dbQueue.Add(dbItemT{fieldName, fieldType, valueInterface, time})
}

func getClickTime(item dbItemT) (timestamp int64) {
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

func dbStore(item dbItemT) {
	//value := reflect.ValueOf(item.ColVal)
	//valueType := value.Type()
	log.Debugf("DB [%d] %s:%s = %v", item.Time, item.ColName, item.ColType, item.ColVal)
	switch item.ColType {
	case "count":
		if counterID, ok := configCounters[item.ColName]; ok {
			database.AddClick(counterID, getClickTime(item))
		} else {
			log.Warn("Unknown counter:", item.ColName)
		}
	case "move":
		log.Warn("MOVE")
	case "led":
		log.Warn("LED")
	case "temp", "humd", "pres":
		database.AddMetric(item.Time, item)
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
	var to = getNearestTimestamp(timestamp, secPeriod)
	var from = to - int64(secPeriod)
	rows := database.popMetricsFrom(from, to)

	row := make(map[string]float64)    // Result
	nowrow := make(map[string]float64) //next old row
	for key, arr := range rows {
		median, err := stats.Median(arr)
		if err == nil {
			nowrow[key] = median
			row[key] = median
		}
	}

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

// resend accomulated data to clickhouse in one row per sec
func dbLoop(ms uint32) {
	dbWg.Add(1)
	defer dbWg.Done()
	ticker := time.NewTicker(time.Millisecond * time.Duration(ms))
	tickerCounter := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	defer tickerCounter.Stop()
	itemChan := dbQueue.getChan()
	for {
		select {
		case item := <-itemChan:
			dbStore(item)
		case <-ticker.C:
			log.Debug("Make clickhouse row")
			dbSaveMetrics(ms)
		case <-tickerCounter.C:
			log.Debug("Make counter row")
			dbProcessEvents()
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
