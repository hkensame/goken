package gormutil

import (
	"database/sql/driver"

	jsoniter "github.com/json-iterator/go"
)

type GormList []string

func (g *GormList) Scan(value interface{}) error {
	return jsoniter.Unmarshal(value.([]byte), &g)
}

func (g GormList) Value() (driver.Value, error) {
	return jsoniter.Marshal(g)
}
