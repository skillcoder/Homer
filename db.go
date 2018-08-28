/* vim: set ts=2 sw=2 sts=2 et: */
package main

import (
//  "strings"
//  "time"
  "reflect"
)

var database map[string]bool = make(map[string]bool)

func dbAdd(field_name string, field_type string, valueInterface interface{}, time int64) {
  value := reflect.ValueOf(valueInterface)
  valueType := value.Type()
  log.Debugf("DB [%s:%s] <%s>=%v", field_name, field_type, valueType, value);
  // Checking whether element type is convertible to function's first argument's type
  //if value.ConvertibleTo(funcType.In(0)) {

}
