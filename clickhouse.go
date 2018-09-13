/* vim: set ts=2 sw=2 sts=2 et: */
package main

import (
  "strings"
  "fmt"
  "github.com/kshvakov/clickhouse"
  "database/sql"
)

// need update by request DDL from clickhouse for table values
var clickhouse_count_of_columns int8 = 10;
var clickhouse_db *sql.DB

func clickhouse_connect() {
  dburl := fmt.Sprintf("tcp://%s:%d?username=%s&password=%s&database=%s&read_timeout=10&write_timeout=20&debug=true",
    config.ClickHouse.Host,
    config.ClickHouse.Port,
    config.ClickHouse.User,
    config.ClickHouse.Pass,
    config.ClickHouse.Name,
  )
  var err error
  clickhouse_db, err = sql.Open("clickhouse", dburl)
	if err != nil {
		log.Fatal(err)
	}
  //defer clickhouse_db.Close()

  // TODO update clickhouse_count_of_columns
  clickhouse_count_of_columns = 6
}

func ch_metric_insert(timestamp int64, row map[string]float64) {
  if err := clickhouse_db.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println(err)
		}
		return
	}

  // result
  sql_fields := make([]string, 0, clickhouse_count_of_columns+1)
  sql_names  := make([]string, 0, clickhouse_count_of_columns+1)
  vals := []interface{}{timestamp}

  for key, val := range row {
    // TODO check key name column exist in our ckickhouse DDL
    // if not - create it by ALTER TABLE values ADD key float8
    log.Debugf("ROW [%s] %.3f", key, val)
    //create insert here like this
    sql_fields = append(sql_fields, fmt.Sprintf("`%s`", key))
    sql_names = append(sql_names, "?")
    vals = append(vals, val)
  }

  if len(sql_names) > 0 {
    //year, month, day := now.Date()
    //weekday := now.Weekday()
    //if weekday == 0 { weekday = 7 }
    //log.Debug(fmt.Sprintf("INSERT INTO `values` SET `ctime` = %d, `year` = %d, `month` = %d, `day` = %d, `weekday` = %d, `hour` = %d, `minute` = %d, %s", timestamp, year, month, day, weekday, now.Hour(), now.Minute(), strings.Join(sql, ", ")))

    var (
      str_sql = "INSERT INTO `metrics` (`ctime`, "+strings.Join(sql_fields, ", ")+") VALUES (?,"+strings.Join(sql_names, ",")+")"
      tx, _ = clickhouse_db.Begin()
      stmt, _ = tx.Prepare(str_sql)
    )
    defer stmt.Close()
    if _, err := stmt.Exec(vals...); err != nil {
      log.Fatal(err)
    }

	  if err := tx.Commit(); err != nil {
		  log.Fatal(err)
  	}
  }
}
