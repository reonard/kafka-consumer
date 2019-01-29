package processor

import "kafka-consumer/db"

func GetDeviceInfo(deviceId int64) (error, DeviceInfo) {

	deviceInfo := DeviceInfo{}

	err := db.MySqlDB.QueryRow("select device_id, device_name, t_device.project as project from t_device "+
		"left join t_customer on project = t_customer.id where device_id = ? ", deviceId).Scan(&deviceInfo.DeviceId,
		&deviceInfo.DeviceName,
		&deviceInfo.Project)

	if err != nil {
		return err, deviceInfo
	}

	return nil, deviceInfo

}
