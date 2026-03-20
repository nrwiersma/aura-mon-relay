package database

// Metric is a single data point to be written to a database.
type Metric struct {
	Measurement string
	Timestamp   int64 // Unix timestamp in seconds.
	Tags        map[string]string
	Fields      map[string]float64
}
