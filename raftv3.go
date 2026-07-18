package raft

// import (
// 	"sync"

// )

// type NodeState int

// const (
// 	Follower NodeState = iota
// 	Leader
// 	Candidate
// )

// type RaftNode struct {
//     mu sync.Mutex

//     id    int
//     peers []string
//     state NodeState

//     currentTerm int
//     votedFor    *int
//     log         []LogEntry

//     // volatile state
//     commitIndex int
//     lastApplied int

//     nextIndex  []int  // nextIndex for each follower,which is the index of the next log entry the leader will send to that follower
//     matchIndex []int  // nextIndex - 1
// }

// type LogEntry struct {
//     Command interface{}
//     Term    int
// }

// type AppendEntries struct{
//     Term int
//     LeaderId int
//     PrevLogIndex int
//     PrevLogTerm int
//     Entries []LogEntry
//     LeaderCommit int

// }

// type AppendEntriesReply struct{
//     Term int
//     Success bool
// }

// type RequestVote struct{
//     Term int
//     CandidateId int
//     LastLogIndex int
//     LastLogTerm int

// }

// type RequestVoteReply struct{
//     Term int
//     VoteGranted bool
// }

// type clientRequest struct {
//     command  interface{}
//     responseCh chan error  // or chan ClientResult
// }

// func (rn *RaftNode) sendRequestVote(peer int, args RequestVote, reply *RequestVoteReply) bool {
//     return true
// }

// func (rn *RaftNode) sendAppendEntries(peer int, args AppendEntries, reply *AppendEntriesReply) bool {
//     return true
// }