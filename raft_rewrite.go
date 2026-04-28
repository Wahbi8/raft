package raft

import "sync"

type NodeState int

const (
	Follower NodeState = iota
	Leader
	Candidate
)

type RaftNode struct {
    mu sync.Mutex

    // identity
    id    int
    peers []string
    state NodeState

    // persistent state (Figure 2)
    currentTerm int
    votedFor    *int // the pointer so it can be null
    log         []LogEntry

    // volatile state (Figure 2)
    commitIndex int
    lastApplied int

    // volatile leader state (Figure 2) — reinitialized after election
    nextIndex  []int
    matchIndex []int
}

type LogEntry struct {
    Command interface{}
    Term    int
}