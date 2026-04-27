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
	IndexId int
	VoteFor int
	HeartBeat chan bool
	Id int
	Log         []LogEntry
    CommitIndex int
    LastApplied int
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

type LogEntry struct {
    Command interface{}
    Term    int
}
type AppendEntriesArg struct{
	Term int
	LeaderId int
	PrevLogIndex LogEntry
	PrevLogTerm int
	//entries
	LeaderCommitIndex int
}
type VotesProcessor struct{ // this a way to compare votes
	Term int
	NodeId []int
}

func (rn *RaftNode) HandleRequestVote(arg RequestVoteArg) RequestVoteResp {
	if arg.Term < rn.TermNumber {
		return RequestVoteResp{Id: rn.Id, Term: rn.TermNumber, vote: false}
	}

	if rn.NodeState = Leader 

	if arg.Term > rn.TermNumber {
		rn.TermNumber = arg.Term
		rn.NodeState = Follower
		rn.VoteFor = 0
	}

	if (rn.VoteFor == 0 || rn.VoteFor == arg.Id) { // incorect need to be fixed
		rn.VoteFor = arg.Id
		// I need to reset the election timer 
		return RequestVoteResp{Id: rn.Id, Term: rn.TermNumber, vote: true}
	}

	return RequestVoteResp{Id: rn.Id, Term: rn.TermNumber, vote: false}
}

func(rn *RaftNode) run(){
	timeoutFollower := time.Duration(150+rand.Intn(150)) * time.Millisecond // the randemazation is not correct	
	timeoutLeader := time.Duration(100) * time.Millisecond

	for {
		switch rn.NodeState {
		case Follower:
            timer := time.NewTimer(timeoutFollower)
			select {
			case <-timer.C:
				rn.NodeState = Candidate
			case <-rn.HeartBeat: // if i am processing an appendEntries from leader there is no need to wait for hartbeat
				timer.Stop()
			// case log := <-rn.Log: 			// To be continued
			}
		case Leader:
            timer := time.NewTimer(timeoutLeader)
			requestVoteArg
			select {
			case <- timer.C:
				//send hartbeat 
				timer.Stop()
				time.Sleep(50 * time.Millisecond)
			case 
			}
		case Candidate:
			rn.TermNumber ++
			//vote for it self somehow 
			rn.VoteFor = rn.Id
			//reset election timer (where the f should i start it???)
			timer := time.NewTimer(timeoutFollower) 
			//send requestvotearg

			requestVoteRespCh := make(chan RequestVoteResp) // this maight not work
			//add the votes to VotesProcessor votes via a pointer
			resp := &VotesProcessor{}
			//find a way to know the number of nodes
			nodesNum := 5
			select{
			case <-timer.C:
				//start new election
			case <-rn.HeartBeat:
				rn.NodeState = Follower
				timer.Stop()
			case vote := <- requestVoteRespCh:
				if vote.vote {
					exists := false

					for _, id := range resp.NodeId {
						if id == vote.Id {
							exists = true
							break
						}
					}

					if !exists {
						resp.NodeId = append(resp.NodeId, vote.Id)
					}
				}

				if len(resp.NodeId) > nodesNum/2 {
					rn.NodeState = Leader
					timer.Stop()
				}
			}
			//if vote received from majority become the leader (how to compare :ah i need the add vote for in RequestVoteResp struct)
			//if received AppendEntriesArg change state to follower
			//if election time out is reached start now election (should i do this with a for or recursion)
			//if became leader and received AppendEnteriesReq and candidate term = leader term candidate acknowledge the curent leader
		}
		
		// how to know if an election time out (i dont i just send the vote it the election is time out it will not recieve the vote	)
	}
}
//TODO: i need to clean up the comments
