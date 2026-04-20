package raft

import (
	"math/rand"
	"time"
)

type NodeState int

const (
	Follower NodeState = iota
	Leader
	Candidate
)

type RaftNode struct {
	NodeState NodeState
	TermNumber int
	ElectionTimer time.Time
	Index int
	VoteFor int
	HeartBeat chan bool
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

func(rn *RaftNode) run(){
	for {
		switch rn.NodeState {
		case Follower:
			timeout := time.Duration(150+rand.Intn(150)) * time.Millisecond
            timer := time.NewTimer(timeout)
			select {
			case <-timer.C:
				rn.NodeState = Candidate
			case <-rn.HeartBeat:
				timer.Stop()
			}
		case Leader:
			timeout := time.Duration(100) * time.Millisecond
            timer := time.NewTimer(timeout)
			select {
			case <- timer.C:
				//send hartbeat 
				timer.Stop()
				time.Sleep(50 * time.Millisecond)
			}
		case Candidate:
			// if candidate send requestvotearg
		}
		
		// how to impliment the random timer for leader election in the split vote case
	}
}
