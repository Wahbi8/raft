package raft 

import (
	"time"
)

type NodeState int

const (
	Leader NodeState = iota
	Follower
	Candidate
)

type RaftNode struct {
	NodeState NodeState
	TermNumber int
	ElectionTimer time.Time
	Index int
	VoteFor int
}

type RequestVoteArg struct{
	Term int
	Id int
}

type RequestVoteResp struct{
	Id int
	Term int
	vote bool
}
type AppendEntriesArg struct{
	// to be filled
}
//to be added: start timer if harthBeat did not arive befor the harthBeat timer, change state and became a condidate

func (rn *RaftNode) HandleRequestVote(arg RequestVoteArg) RequestVoteResp {
	if arg.Term < rn.TermNumber {
		return RequestVoteResp{Id: rn.Index, Term: rn.TermNumber, vote: false}
	}

	if arg.Term > rn.TermNumber {
		rn.TermNumber = arg.Term
		rn.NodeState = Follower
		rn.VoteFor = 0 // Reset my vote for this new term
	}

	if (rn.VoteFor == 0 || rn.VoteFor == arg.Id) {
		rn.VoteFor = arg.Id
		// Every time I give a vote, I should reset my election timer 
		// I need to create function rn.resetElectionTimer() 
		return RequestVoteResp{Id: rn.Index, Term: rn.TermNumber, vote: true}
	}

	return RequestVoteResp{Id: rn.Index, Term: rn.TermNumber, vote: false}
}

// create a goroutine to send harthbeat if the node is a leader 
// create a goroutine if the node is a follower to become a candidate
go func() {
	for {
		// how to decide the state of the node
		if 
		// start timer
		// if leader send hartbeat and reset timer
		// if follower reset timer if a harthbeat recieved or became a candidate if reached time limit 
		// if candidate send requestvotearg
		// how to impliment the random timer for leader election in the split vote case
	}
}()
