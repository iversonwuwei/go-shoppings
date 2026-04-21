package utils

import (
	"errors"
	"sync"
	"time"
)

// Snowflake: 1 位符号 + 41 位时间戳(ms) + 10 位机器 ID + 12 位序列
type Snowflake struct {
	mu       sync.Mutex
	epoch    int64
	nodeID   int64
	lastTS   int64
	sequence int64
}

const (
	snowEpoch       = int64(1704067200000) // 2024-01-01
	nodeBits        = 10
	seqBits         = 12
	maxNode         = -1 ^ (-1 << nodeBits)
	maxSeq    int64 = -1 ^ (-1 << seqBits)
	timeShift       = nodeBits + seqBits
	nodeShift       = seqBits
)

func NewSnowflake(nodeID int64) (*Snowflake, error) {
	if nodeID < 0 || nodeID > maxNode {
		return nil, errors.New("node id out of range")
	}
	return &Snowflake{epoch: snowEpoch, nodeID: nodeID}, nil
}

func (s *Snowflake) NextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UnixMilli()
	if now == s.lastTS {
		s.sequence = (s.sequence + 1) & maxSeq
		if s.sequence == 0 {
			for now <= s.lastTS {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		s.sequence = 0
	}
	s.lastTS = now
	return ((now - s.epoch) << timeShift) | (s.nodeID << nodeShift) | s.sequence
}

var defaultNode, _ = NewSnowflake(1)

func NextID() int64 { return defaultNode.NextID() }
