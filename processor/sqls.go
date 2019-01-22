package processor

import "kafka-consumer/db"

func GetDeviceInfo(deviceid int64) (error, DeviceInfo) {

	deviceInfo := DeviceInfo{}

	err := db.MySqlDB.QueryRow("select device_id, device_name, device.project as project, customer.name as customer from device "+
		"left join project on device.project = project.name "+
		"left join customer on project.customer = customer.name where device_id = ?", deviceid).Scan(&deviceInfo.DeviceId,
		&deviceInfo.DeviceName,
		&deviceInfo.Project,
		&deviceInfo.Customer)
	if err != nil {
		return err, deviceInfo
	}

	return nil, deviceInfo

}
