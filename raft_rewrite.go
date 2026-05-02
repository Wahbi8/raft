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

type AppendEntries struct{
    Term int
    LeaderId int
    PrevLogIndex int
    PrevLogTerm int
    Entries []LogEntry
    LeaderCommit int

}

type AppendEntriesReply struct{
    Term int
    Success bool
}

type RequestVote struct{
    Term int
    CandidateId int
    LastLogIndex int
    LastLogTerm int

}

type RequestVoteReply struct{
    Term int
    VoteGranted bool
}

func (rn *RaftNode) sendRequestVote(peer int, args RequestVote, reply *RequestVoteReply) bool {
    return true
}

func (rn *RaftNode) sendAppendEntries(peer int, args AppendEntries, reply *AppendEntriesReply) bool {
    return true
}

func (rn *RaftNode) HandleRequestVote(arg RequestVote) RequestVoteReply {
	if arg.Term < rn.currentTerm {
		return RequestVoteReply{Term: rn.currentTerm, VoteGranted: false}
	}

	if arg.Term > rn.currentTerm {
		rn.currentTerm = arg.Term
		rn.state = Follower
		rn.votedFor = nil
	}

    lastLogIndex := 0
    lastLogTerm := 0

    if len(rn.log) > 0 {
        lastLogIndex = len(rn.log) - 1
        lastLogTerm = rn.log[lastLogIndex].Term
    }

    logOk := arg.LastLogTerm > lastLogTerm ||
        (arg.LastLogTerm == lastLogTerm && arg.LastLogIndex >= lastLogIndex) 

    // logOk is to make sure that the candidates have at least the same log as the follower to prevent electing a leader with less logs
    if (rn.votedFor == nil || *rn.votedFor == arg.CandidateId) && logOk {
        rn.votedFor = &arg.CandidateId
	    return RequestVoteReply{ Term: rn.currentTerm, VoteGranted: true}
    }
	return RequestVoteReply{ Term: rn.currentTerm, VoteGranted: false}
}

func (rn *RaftNode) HandleAppendEntries(arg AppendEntries) AppendEntriesReply {
    if rn.currentTerm > arg.Term || arg.PrevLogIndex != arg.PrevLogTerm {
        return AppendEntriesReply{Term: rn.currentTerm, Success: false }
    }

    
    return AppendEntriesReply{}
}