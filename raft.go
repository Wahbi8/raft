package raft 

import (
	"time"
)

type NodeState int

const (
	Leader NodeState = iota
	Follower
	Condidate
)

type RaftNode struct {
	NodeState NodeState
	TermNumber int
	CurrentTerm int
	ElectionTimeLimit time.Time
	Index int
	Vote map[string]string

}
