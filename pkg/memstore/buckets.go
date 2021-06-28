package memstore

type BucketName int

const (
	// TelemetryBucket is the name of the store bucket used by the telemetry service
	TelemetryBucket BucketName = iota + 1
)
