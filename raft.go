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
		// I need to create function rn.setElectionTimer() 
		return RequestVoteResp{Id: rn.Index, Term: rn.TermNumber, vote: true}
	}

	return RequestVoteResp{Id: rn.Index, Term: rn.TermNumber, vote: false}
}

func(rn *RaftNode) run(){
	timeoutFollower := time.Duration(150+rand.Intn(150)) * time.Millisecond
	timeoutLeader := time.Duration(100) * time.Millisecond

	for {
		switch rn.NodeState {
		case Follower:
            timer := time.NewTimer(timeoutFollower)
			select {
			case <-timer.C:
				rn.NodeState = Candidate
			case <-rn.HeartBeat:
				timer.Stop()
			}
		case Leader:
            timer := time.NewTimer(timeoutLeader)
			select {
			case <- timer.C:
				//send hartbeat 
				timer.Stop()
				time.Sleep(50 * time.Millisecond)
			}
		case Candidate:
			rn.TermNumber ++
			//vote for it self somehow
			//reset election timer (where the f should i start it???)
			//send requestvotearg
			//if vote received from majority become the leader (how to compare :ah i need the add vote for in RequestVoteResp struct)
			//if received AppendEntriesArg change state to follower
			//if election time out is reached start now election (should i do this with a for)
		}
		
		// how to know if an election time out (i dont i just send the vote it the election is time out it will not recieve the vote	)
	}
}
