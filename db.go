/* vim: set ts=2 sw=2 sts=2 et: */
package main

import (
  "time"
//  "reflect"
  "sync"
  "sort"
  "github.com/montanaflynn/stats"
)

type dbItem_t struct {
  ColName string
  ColType string
  ColVal float64
  Time int64
}

type dbQueue_t struct {
  ch chan dbItem_t
}

func (q *dbQueue_t) Add(item dbItem_t) {
  q.ch <- item
}

func (q *dbQueue_t) getChan() <-chan dbItem_t {
  return q.ch
}

type db_t struct {
  m map[int64][]dbItem_t
}

func NewDatabase() *db_t {
  return &db_t{
    m: make(map[int64][]dbItem_t),
  }
}

func (c *db_t) Load(key int64) ([]dbItem_t, bool) {
  val, ok := c.m[key]
  return val, ok
}

func (c *db_t) Add(key int64, item dbItem_t) {
  column, ok := c.Load(key)
  if !ok {
    column = make([]dbItem_t, 0, clickhouse_count_of_columns+1)
  }

  column = append(column, item)
  c.m[key] = column
}


var dbQueue dbQueue_t
var database *db_t
var dbShutdownChan chan bool = make(chan bool)
var dbWg sync.WaitGroup
var oldrow map[string]float64 = make(map[string]float64)

func init() {
  database = NewDatabase()
  dbQueue = dbQueue_t{
    ch:make(chan dbItem_t, 128),
  }
}

func dbAddMetric(field_name string, field_type string, valueInterface float64, time int64) {
  dbQueue.Add(dbItem_t{field_name, field_type, valueInterface, time})
}

func dbAddEvent(field_name string, field_type string, valueInterface uint64, time int64) {
  //dbQueue.Add(dbItem_t{field_name, field_type, valueInterface, time})
}


func dbStore(item dbItem_t) {
  //value := reflect.ValueOf(item.ColVal)
  //valueType := value.Type()
  log.Debugf("DB [%s:%s] <%T>=%v", item.ColName, item.ColType, item.ColVal, item.ColVal)
  database.Add(item.Time, item)
}

func dbDoTransfer() {
  now := time.Now()
  var timestamp int64 = now.Unix()
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
        row_key := item.ColName+":"+item.ColType
        rows[row_key] = append(rows[row_key], item.ColVal)
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
    median, _ := stats.Median(arr)
    nowrow[key] = median
    row[key] = median
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

  go ch_metric_insert(timestamp, row);
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
