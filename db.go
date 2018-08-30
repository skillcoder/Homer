/* vim: set ts=2 sw=2 sts=2 et: */
package main

import (
//  "strings"
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
    column = make([]dbItem_t, 0, 10)
  }

  column = append(column, item)
  c.m[key] = column
}


var dbQueue dbQueue_t
var database *db_t
var dbShutdownChan chan bool = make(chan bool)
var dbWg sync.WaitGroup

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
  var timestamp int64 = time.Now().Unix()
  from := timestamp - 5;
  keys := make([]int64, 0, len(database.m))
  for k := range database.m {
    keys = append(keys, k)
  }

  sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
  row := make(map[string][]float64)
  for _, key := range keys {
    if key >= from {
      //log.Debugf("key [%d] %v", key, item);
      for _, item := range database.m[key] {
        row_key := item.ColName+":"+item.ColType
        row[row_key] = append(row[row_key], item.ColVal)
      }
      delete(database.m, key);
    } else {
      log.Warnf("key [%d] %v", key, database.m[key]);
    }
  }

  for key, arr := range row {
    median, _ := stats.Median(arr)
    log.Debugf("ROW [%s] %.3f", key, median);
  }
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
      log.Debug("Ticker ticked")
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
