package db

import (
	"kafka-consumer/lib"
	"testing"
)

func TestMySql(t *testing.T) {

	lib.InitCfg()
	InitMySQLDB(lib.AppCfg.MySqlURL)
	//insertData()
	querymyData()

}

func querymyData() {

	ExecuteUpdate(
		"update device set device_status = ? where device_id = ?",
		"alarm",
		"device1")
}
