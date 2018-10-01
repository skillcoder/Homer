package main

/* vim: set ts=2 sw=2 sts=2 ff=unix ft=go noet: */

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/kshvakov/clickhouse"
)

// need update by request DDL from clickhouse for table values
var clickhouseMetricCount = 12
var clickhouseMetrics = make(map[string]struct{})
var clickhouseDb *sql.DB
var clickhouseCtx = context.Background()

func clickhouseCheckFieldName(name string) (result bool) {
	re := regexp.MustCompile(`^[^a-zA-Z0-9:\-_]+$`)
	result = re.MatchString(name)
	return !result
}

func clickhouseAddMetric(fieldName, fieldType string) {
	if fieldType == "Nullable(Float32)" {
		clickhouseMetrics[fieldName] = struct{}{}
		clickhouseMetricCount = len(clickhouseMetrics)
		log.Info("Metric init ", fieldName)
	} else {
		verbosePrint("Ignore [" + fieldName + "] cuz fieldType = " + fieldType)
	}
}

// Get clickhouseMetricCount from real batabase table structure
func clickhouseInitMetrics() {
	rows, err := clickhouseDb.QueryContext(clickhouseCtx, "DESCRIBE TABLE `metrics`")
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		e := rows.Close()
		if e != nil {
			log.Fatal(e)
		}
	}()

	cols, err := rows.Columns()
	if err != nil {
		log.Fatal("ClickHouse get columns ERR:", err)
	}

	colCount := len(cols)
	verbosePrint(fmt.Sprintf("DESCRIBE Cols: %d", colCount))
	vals := make([]string, colCount)
	dist := make([]interface{}, colCount)
	for i := range cols {
		dist[i] = &vals[i]
	}

	for rows.Next() {
		if err := rows.Scan(dist...); err != nil {
			log.Fatal(err)
		}

		if dist[0] != nil && dist[1] != nil {
			clickhouseAddMetric(vals[0], vals[1])
		}
	}

	verbosePrint(fmt.Sprintf("clickhouseMetricCount = %d", clickhouseMetricCount))
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func clickhouseConnect() {
	dburl := fmt.Sprintf("tcp://%s:%d?username=%s&password=%s&database=%s&read_timeout=10&write_timeout=20&debug=true",
		config.ClickHouse.Host,
		config.ClickHouse.Port,
		config.ClickHouse.User,
		config.ClickHouse.Pass,
		config.ClickHouse.Name,
	)
	var err error
	clickhouseDb, err = sql.Open("clickhouse", dburl)
	if err != nil {
		log.Fatal(err)
	}
	//defer clickhouseDb.Close()

	clickhouseInitMetrics()
}

func clickhousePing() {
	if err := clickhouseDb.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println(err)
		}
		return
	}
}

func clickhouseMetricInsert(timestamp int64, row map[string]float64) {
	clickhousePing()

	// result
	sqlFields := make([]string, 0, clickhouseMetricCount+1)
	sqlNames := make([]string, 0, clickhouseMetricCount+1)
	vals := []interface{}{timestamp}

	for key, val := range row {
		// check key name column exist in our ckickhouse DDL
		if _, ok := clickhouseMetrics[key]; !ok {
			// if not exist - create it by ALTER TABLE values ADD key float8
			if ok := clickhouseCheckFieldName(key); ok {
				_, err := clickhouseDb.Exec("ALTER TABLE `metrics` ADD COLUMN `$1` Nullable(Float32)", key)
				if err != nil {
					log.Errorf("Cant ADD COLUMN [%s] to database: %v", key, err)
					continue
				}

				clickhouseAddMetric(key, "Nullable(Float32)")
			} else {
				log.Warnf("Invalid COLUMN name [%s]", key)
				continue
			}
		}

		log.Debugf("ROW [%s] %.3f", key, val)
		//create insert here like this
		sqlFields = append(sqlFields, fmt.Sprintf("`%s`", key))
		sqlNames = append(sqlNames, "?")
		vals = append(vals, val)
	}

	if len(sqlNames) > 0 {
		/*
		   year, month, day := now.Date()
		   weekday := now.Weekday()
		   if weekday == 0 { weekday = 7 }
		   log.Debug(fmt.Sprintf("""
		     INSERT INTO `values` SET
		       `ctime` = %d,
		       `year` = %d,
		       `month` = %d,
		       `day` = %d,
		       `weekday` = %d,
		       `hour` = %d,
		       `minute` = %d,
		       %s""",timestamp, year, month, day, weekday, now.Hour(), now.Minute(), strings.Join(sql, ", ")))
		*/

		var (
			strSQL = "INSERT INTO `metrics`" +
				" (`ctime`, " + strings.Join(sqlFields, ", ") + ")" +
				" VALUES (?," + strings.Join(sqlNames, ",") + ")"
			tx, _   = clickhouseDb.Begin()
			stmt, _ = tx.Prepare(strSQL) // nolint: safesql
		)
		//defer tx.Rollback()
		defer func() {
			err := stmt.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()
		if _, err := stmt.Exec(vals...); err != nil {
			log.Fatal(err)
		}

		if err := tx.Commit(); err != nil {
			log.Fatal(err)
		}
	}
}
