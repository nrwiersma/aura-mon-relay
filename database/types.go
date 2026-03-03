package database

// Metric is a single data point that can be stored in the database.
type Metric struct {
	Measurement string
	Timestamp   int64
	Tags        map[string]string
	Fields      map[string]float64
}
