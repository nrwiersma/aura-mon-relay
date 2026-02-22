package database

type Metric struct {
	Measurement string
	Timestamp   int64
	Tags        map[string]string
	Fields      map[string]float64
}
