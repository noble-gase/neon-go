package closekit

type Priority int

const (
	P0 Priority = iota
	P1
	P2
	P3
	P4
	P5
	P6
	P7  // MQ
	P8  // WorkerPool
	P9  // Redis
	P10 // DB
)
