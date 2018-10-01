package database

/* vim: set ts=2 sw=2 sts=2 ff=unix ft=go noet: */

import (
	"fmt"
	"sort"
	"strconv"
)

// StatT defines HTTP API response giving database stats
type StatT struct {
	MetricCount int `json:"metric_count"`
	ClickCount  int `json:"click_count"`
}

// DbItemT is struct of one Metric/Event item {ColName, ColType, ColVal, Time}
type DbItemT struct {
	ColName string
	ColType string
	ColVal  string
	Time    int64
}

// CreateItem just helper for create item of type DbItemT
func CreateItem(fieldName string, fieldType string, valueInterface string, time int64) DbItemT {
	return DbItemT{
		fieldName,
		fieldType,
		valueInterface,
		time,
	}
}

// DbQueueT queue type contain chanel for send to database data from any gorutines
type DbQueueT struct {
	ch chan DbItemT
}

// NewQueue constructor of DbQueueT
func NewQueue(queueSize int) *DbQueueT {
	return &DbQueueT{
		ch: make(chan DbItemT, queueSize),
	}
}

// Add - send item(Metric/Event) to database
func (q *DbQueueT) Add(item DbItemT) {
	q.ch <- item
}

// GetChan get chanel for processing added to database queue data
func (q *DbQueueT) GetChan() <-chan DbItemT {
	return q.ch
}

// DbT is Database class for store metrics and events like counter clicks
type DbT struct {
	realMetricCount int
	m               map[int64][]DbItemT
	clicks          map[uint8][]int64
}

// New - create instance of DbT database
func New(realMetricCount int) *DbT {
	return &DbT{
		realMetricCount: realMetricCount,
		m:               make(map[int64][]DbItemT),
		clicks:          make(map[uint8][]int64),
	}
}

func (c *DbT) load(key int64) ([]DbItemT, bool) {
	val, ok := c.m[key]
	return val, ok
}

// AddMetric - store metric in database
func (c *DbT) AddMetric(key int64, item DbItemT) {
	column, ok := c.load(key)
	if !ok {
		column = make([]DbItemT, 0, c.realMetricCount+1)
	}

	column = append(column, item)
	c.m[key] = column
}

// AddClick - store click in database
func (c *DbT) AddClick(key uint8, time int64) {
	click, ok := c.clicks[key]
	if !ok {
		click = make([]int64, 0, 10)
	}

	click = append(click, time)
	c.clicks[key] = click
}

// PopMetricsFrom agreagate metrics from database (with delete it) in time range [from, to)
func (c *DbT) PopMetricsFrom(from, to int64) (rows map[string][]float64) {
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
						fmt.Printf("[%d] %s %s %s convert to float: %s", item.Time, item.ColName, item.ColType, item.ColVal, err)
						continue
					}

					rows[rowKey] = append(rows[rowKey], value)
				}
				delete(c.m, key)
			} else {
				// FIXME need send it too (its late data)
				fmt.Printf("key [%d] %v", key, c.m[key])
			}
		}
	}

	return rows
}

// GetStat get current stat of database
func (c *DbT) GetStat() StatT {
	return StatT{
		len(c.m),
		len(c.clicks),
	}
}
