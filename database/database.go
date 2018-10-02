package database

/* vim: set ts=2 sw=2 sts=2 ff=unix ft=go noet: */

import (
	"fmt"
	//"sort"
	"strconv"

	"github.com/skillcoder/homer/helper"
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

	click = helper.InsertSortedInt64(click, time)
	c.clicks[key] = click
}

// PopMetricsFrom agreagate metrics from database (with delete it) in time range [from, to)
func (c *DbT) PopMetricsFrom(from, to int64) (rows map[string][]float64, isHaveOlder bool) {
	keys := make([]int64, 0, len(c.m))
	for k := range c.m {
		keys = append(keys, k)
	}

	//sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
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
				// FIXME remove it
				fmt.Printf("older key [%d] %v", key, c.m[key])
				isHaveOlder = true
			}
		}
	}

	return rows, isHaveOlder
}

// PopClicksFrom agreagate clicks from database (with delete it) in time range [from, to)
func (c *DbT) PopClicksFrom(from, to int64) (rows map[uint8]uint16, isHaveOlder bool) {
	rows = make(map[uint8]uint16)
	for counterID, data := range c.clicks {
		rows[counterID] = 0
		older := make([]int64, 0)
		for _, t := range data {
			if t < to {
				c.clicks[counterID] = c.clicks[counterID][1:] // delete first element
				if t >= from {
					rows[counterID]++
				} else {
					isHaveOlder = true
					older = append(older, t)
					// FIXME remove it
					fmt.Printf("older click [%d] %d", counterID, t)
				}
			} else {
				break
			}
		}

		if len(older) > 0 {
			newerCount := len(c.clicks[counterID])
			newer := make([]int64, newerCount)
			if newerCount > 0 {
				copy(newer, c.clicks[counterID])
			}

			c.clicks[counterID] = make([]int64, newerCount+len(older))
			copy(c.clicks[counterID], older)

			if newerCount > 0 {
				copy(c.clicks[counterID][len(older)-1:], newer)
			}
		}
	}

	return rows, isHaveOlder
}

// GetStat get current stat of database
func (c *DbT) GetStat() StatT {
	var metricCount = 0
	for _, v := range c.m {
		metricCount += len(v)
	}

	var clickCount = 0
	for _, v := range c.clicks {
		clickCount += len(v)
	}

	return StatT{
		metricCount,
		clickCount,
	}
}
