package main
/* vim: set ts=2 sw=2 sts=2 et: */

import (
  "strings"
  "fmt"
  "database/sql"

  "github.com/kshvakov/clickhouse"
)

// need update by request DDL from clickhouse for table values
var clickhouseCountOfColumns int8 = 10;
var clickhouseDb *sql.DB

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

  // TODO update clickhouseCountOfColumns
  clickhouseCountOfColumns = 6
}

func clickhouseMetricInsert(timestamp int64, row map[string]float64) {
  if err := clickhouseDb.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println(err)
		}
		return
	}

  // result
  sqlFields := make([]string, 0, clickhouseCountOfColumns+1)
  sqlNames  := make([]string, 0, clickhouseCountOfColumns+1)
  vals := []interface{}{timestamp}

  for key, val := range row {
    // TODO check key name column exist in our ckickhouse DDL
    // if not - create it by ALTER TABLE values ADD key float8
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
      strSQL = "INSERT INTO `metrics`"+
        " (`ctime`, "+strings.Join(sqlFields, ", ")+")"+
        " VALUES (?,"+strings.Join(sqlNames, ",")+")"
      tx, _ = clickhouseDb.Begin()
      stmt, _ = tx.Prepare(strSQL)
    )
    //defer tx.Rollback()
    defer checkDefer(stmt.Close())
    if _, err := stmt.Exec(vals...); err != nil {
      log.Fatal(err)
    }

	  if err := tx.Commit(); err != nil {
		  log.Fatal(err)
  	}
  }
}
