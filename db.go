package main
/* vim: set ts=2 sw=2 sts=2 et: */

import (
  "time"
//  "reflect"
  "sync"
  "sort"

  "github.com/montanaflynn/stats"
)

type dbItemT struct {
  ColName string
  ColType string
  ColVal float64
  Time int64
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
  m map[int64][]dbItemT
}

func newDatabase() *dbT {
  return &dbT{
    m: make(map[int64][]dbItemT),
  }
}

func (c *dbT) Load(key int64) ([]dbItemT, bool) {
  val, ok := c.m[key]
  return val, ok
}

func (c *dbT) Add(key int64, item dbItemT) {
  column, ok := c.Load(key)
  if !ok {
    column = make([]dbItemT, 0, clickhouseCountOfColumns+1)
  }

  column = append(column, item)
  c.m[key] = column
}


var dbQueue dbQueueT
var database *dbT
var dbShutdownChan = make(chan bool)
var dbWg sync.WaitGroup
var oldrow = make(map[string]float64)

func init() {
  database = newDatabase()
  dbQueue = dbQueueT{
    ch:make(chan dbItemT, 128),
  }
}

func dbAddMetric(fieldName string, fieldType string, valueInterface float64, time int64) {
  dbQueue.Add(dbItemT{fieldName, fieldType, valueInterface, time})
}

func dbAddEvent(fieldName string, fieldType string, valueInterface uint64, time int64) {
  //dbQueue.Add(dbItemT{fieldName, fieldType, valueInterface, time})
}


func dbStore(item dbItemT) {
  //value := reflect.ValueOf(item.ColVal)
  //valueType := value.Type()
  log.Debugf("DB [%s:%s] <%T>=%v", item.ColName, item.ColType, item.ColVal, item.ColVal)
  database.Add(item.Time, item)
}

func dbDoTransfer() {
  now := time.Now()
  var timestamp = now.Unix()
  from := timestamp - 5;
  keys := make([]int64, 0, len(database.m))
  for k := range database.m {
    keys = append(keys, k)
  }

  sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
  rows := make(map[string][]float64)
  for _, key := range keys {
    if key >= from {
      //log.Debugf("key [%d] %v", key, item);
      for _, item := range database.m[key] {
        rowKey := item.ColName+":"+item.ColType
        rows[rowKey] = append(rows[rowKey], item.ColVal)
      }
      delete(database.m, key);
    } else {
      // FIXME need send it too (its late data)
      log.Warnf("key [%d] %v", key, database.m[key]);
    }
  }

  row := make(map[string]float64) // Result
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

  go clickhouseMetricInsert(timestamp, row);
}

// resend accomulated data to clickhouse in one row per sec
func dbLoop(ms uint32) {
  dbWg.Add(1)
  defer dbWg.Done()
  ticker := time.NewTicker(time.Millisecond * time.Duration(ms))
  defer ticker.Stop()
  itemChan := dbQueue.getChan()
  for {
    select {
    case item := <- itemChan:
      dbStore(item)
    case <- ticker.C:
      log.Debug("Make clickhouse row")
      dbDoTransfer()
    case <- dbShutdownChan:
      log.Debug("DB shuting down")
      dbDoTransfer()
      return
    }
  }
}

func dbShutdown() {
  dbShutdownChan <- true
  dbWg.Wait()
  log.Info("DB shutdowned")
}
