// +build !nosqlite

package cmd

import (
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)
