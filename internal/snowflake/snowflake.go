package snowflake


import (
	"fmt"
	"sync"
	"time"
)

const (
	epoch             int64 = 1609459200000 // 2021-01-01 00:00:00 毫秒时间戳作为基准,这个事件可以自己换
	timestampBits     uint  = 41
	datacenterIDBits  uint  = 5
	workerIDBits      uint  = 5
	sequenceBits      uint  = 12
	maxDatacenterID   int64 = -1 ^ (-1 << datacenterIDBits)
	maxWorkerID       int64 = -1 ^ (-1 << workerIDBits)
	maxSequence       int64 = -1 ^ (-1 << sequenceBits)
	workerIDShift     uint  = sequenceBits
	datacenterIDShift uint  = sequenceBits + workerIDBits
	timestampShift    uint  = sequenceBits + workerIDBits + datacenterIDBits
)

type Snowflake struct {
	mu           sync.Mutex
	lastTimestamp int64
	datacenterID  int64
	workerID      int64
	sequence      int64
}

// NewSnowflake 新建一个雪花算法实例
func NewSnowflake(datacenterID, workerID int64) (*Snowflake, error) {
	if datacenterID > maxDatacenterID || datacenterID < 0 {
		return nil, fmt.Errorf("datacenter ID must be between 0 and %d", maxDatacenterID)
	}
	if workerID > maxWorkerID || workerID < 0 {
		return nil, fmt.Errorf("worker ID must be between 0 and %d", maxWorkerID)
	}
	return &Snowflake{
		datacenterID: datacenterID,
		workerID:     workerID,
		sequence:     0,
		lastTimestamp: -1,
	}, nil
}

// Generate 生成唯一 ID
func (s *Snowflake) Generate() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := timeGen()
	if timestamp == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			timestamp = tilNextMillis(s.lastTimestamp)
		}
	} else {
		s.sequence = 0
	}
	s.lastTimestamp = timestamp

	return ((timestamp - epoch) << timestampShift) |
		(s.datacenterID << datacenterIDShift) |
		(s.workerID << workerIDShift) |
		s.sequence
}

// tillNextMillis 等待下一毫秒
func tilNextMillis(lastTimestamp int64) int64 {
	timestamp := timeGen()
	for timestamp <= lastTimestamp {
		timestamp = timeGen()
	}
	return timestamp
}

// timeGen 获取当前时间戳
func timeGen() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

var GlabolSF *Snowflake

// // 演示用
// func main() {
// 	sf, err := NewSnowflake(1, 1)
// 	if err != nil {
// 		panic(err)
// 	}
// 	for i := 0; i < 10; i++ {
// 		fmt.Println(sf.Generate())
// 	}
// }
