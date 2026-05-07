package raft

import (
    "sync"
    "time"
    "math/rand"
)
    
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

    // i am not sure this is correct (but it is fixing a problem)
    AppendEntriesCh chan AppendEntries
    RequestVoteCh chan RequestVote

    RequestVoteReplyCh chan []RequestVoteReply
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
    if arg.Term < rn.currentTerm {
        return AppendEntriesReply{Term: rn.currentTerm, Success: false}
    }

    if arg.Term > rn.currentTerm {
        rn.currentTerm = arg.Term
        rn.state = Follower
        rn.votedFor = nil
    }

    if arg.PrevLogIndex >= len(rn.log) || rn.log[arg.PrevLogIndex].Term != arg.PrevLogTerm {
        return AppendEntriesReply{Term: rn.currentTerm, Success: false}
    }

    // Rule 3 from paper
    for i, entry := range arg.Entries {
        logIndex := arg.PrevLogIndex + 1 + i
        if logIndex < len(rn.log) && rn.log[logIndex].Term != entry.Term {
            rn.log = rn.log[:logIndex] // delete from conflict point onwards
            break
        }
    }

    // Rule 4 form paper
    for i, entry := range arg.Entries {
        logIndex := arg.PrevLogIndex + 1 + i
        if logIndex > len(rn.log) {
            rn.log = append(rn.log, entry)
        }
    }

    // Rule 5 — update commitIndex
    if arg.LeaderCommit > rn.commitIndex {
        rn.commitIndex = min(arg.LeaderCommit, len(rn.log)-1)
    }

    return AppendEntriesReply{Term: rn.currentTerm, Success: true}
}

func (rn *RaftNode) run() {
    leaderTime := time.Duration(100) * time.Millisecond
    randomTime := time.Duration(150+rand.Intn(150)) * time.Millisecond // the randemazation is not correct	
  
    leaderTimer := time.NewTimer(leaderTime)
    randomTimer := time.NewTimer(randomTime)

    nodeNum := 5
    for {
        switch rn.state {
        case Follower:
            select{
            case <-randomTimer.C:
                rn.state = Candidate
                randomTimer.Reset(randomTime)
            case arg := <- rn.AppendEntriesCh:
                randomTimer.Reset(randomTime)
                rn.HandleAppendEntries(arg)
            case arg := <- rn.RequestVoteCh:
                rn.HandleRequestVote(arg)
            }
        case Candidate:
            rn.currentTerm ++
            rn.votedFor = &rn.id
            randomTimer.Reset(randomTime)

            logNum := 0

            if len(rn.log) == 0 {
                logNum = 1
            } else {
                logNum = len(rn.log) - 1
            }

            argRV := RequestVote{
                Term: rn.currentTerm,
                CandidateId: rn.id,
                LastLogIndex: logNum,
                LastLogTerm: rn.log[logNum].Term , 
            }

            // i need a finction that allow me to know how many nodes are available      
            // i will need this information in 3 spret places (to send the vote and append requests and to calculate the votes)
            // for now i will fill the node number statically (and i should use rn.peers )
            
            votes := 1

            replyCh := make(chan RequestVoteReply, nodeNum-1) // i should use len(rn.peers) insted of static number

            for num := 1; num < nodeNum; num++ {
                go func(peer int) {
                    argRVRepply := &RequestVoteReply{}
                    ok := rn.sendRequestVote(peer, argRV, argRVRepply)
                    if ok {
                        replyCh <- *argRVRepply
                    }
                }(num)
            }
            ElectionLoop:
            for {
                select{
                case reply := <-replyCh:
                if reply.VoteGranted {
                    votes++
                    if votes >= nodeNum / 2 {
                        rn.state = Leader
                        // initialize leader state, send heartbeats...
                        break ElectionLoop
                    }
                }
                case arg := <- rn.AppendEntriesCh:
                    rn.state = Follower
                    randomTimer.Reset(randomTime)
                    rn.HandleAppendEntries(arg)
                    break ElectionLoop
                case <-randomTimer.C:
                    break ElectionLoop
                }                
            }
        
        case Leader:
            select{
            case <-leaderTimer.C:
                arg := AppendEntries{}
                argRepply := AppendEntriesReply{}
                for num := 1; num < nodeNum; num++{
                    rn.sendAppendEntries(num, arg, &argRepply)
                }
                leaderTimer.Reset(leaderTime)
            
            }
        }
    }
}