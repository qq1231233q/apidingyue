package system_setting

import "os"

var ServerAddress = "http://localhost:3000"
var WorkerUrl = ""
var WorkerValidKey = ""
var InternalSyncSecret = os.Getenv("INTERNAL_SYNC_SECRET")
var WorkerAllowHttpImageRequestEnabled = false

func EnableWorker() bool {
	return WorkerUrl != ""
}
