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
	ElectionTimer *time.Time
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
//to be fixed: where i fiil the value of the current node
var currentNode = RaftNode{
	NodeState: 2,
	TermNumber: 1,
	Index: 1,
	VoteFor: 0,
}

//to be added: if harthBeat did not arive befor the harthBeat timer, change state and became a condidate

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
