package raft

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

type NodeState int

const (
	Follower NodeState = iota
	Leader
	Candidate
)

type RaftNode struct {
	mu sync.Mutex

	id    int
	peers []string
	state NodeState

	currentTerm int
	votedFor    *int // pointer so it can be null
	log         []LogEntry

	// volatile state
	commitIndex int
	lastApplied int

	// volatile leader state (Figure 2) — reinitialized after election
	nextIndex  []int // index of the next log entry the leader will send to that follower
	matchIndex []int

	// Simulated RPC channels
	AppendEntriesCh chan AppendEntries
	RequestVoteCh   chan RequestVote

	ClientCommandCh chan clientRequest
}

type LogEntry struct {
	Command interface{}
	Term    int
}

type AppendEntries struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

type RequestVote struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

type clientRequest struct {
	command    interface{}
	responseCh chan error
}

func (rn *RaftNode) sendRequestVote(peer int, args RequestVote, reply *RequestVoteReply) bool {
	return true
}

func (rn *RaftNode) sendAppendEntries(peer int, args AppendEntries, reply *AppendEntriesReply) bool {
	return true
}

func (rn *RaftNode) HandleRequestVote(arg RequestVote) RequestVoteReply {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if arg.Term < rn.currentTerm {
		return RequestVoteReply{Term: rn.currentTerm, VoteGranted: false}
	}

	if arg.Term > rn.currentTerm {
		rn.currentTerm = arg.Term
		rn.state = Follower
		rn.votedFor = nil
	}

	lastLogIndex := len(rn.log) - 1
	lastLogTerm := rn.log[lastLogIndex].Term

	logOk := arg.LastLogTerm > lastLogTerm ||
		(arg.LastLogTerm == lastLogTerm && arg.LastLogIndex >= lastLogIndex)

	if (rn.votedFor == nil || *rn.votedFor == arg.CandidateId) && logOk {
		rn.votedFor = &arg.CandidateId
		return RequestVoteReply{Term: rn.currentTerm, VoteGranted: true}
	}

	return RequestVoteReply{Term: rn.currentTerm, VoteGranted: false}
}

func (rn *RaftNode) HandleAppendEntries(arg AppendEntries) AppendEntriesReply {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if arg.Term < rn.currentTerm {
		return AppendEntriesReply{Term: rn.currentTerm, Success: false}
	}

	if arg.Term > rn.currentTerm || rn.state == Candidate {
		rn.currentTerm = arg.Term
		rn.state = Follower
		rn.votedFor = nil
	}

	if arg.PrevLogIndex >= len(rn.log) || rn.log[arg.PrevLogIndex].Term != arg.PrevLogTerm {
		return AppendEntriesReply{Term: rn.currentTerm, Success: false}
	}

	for i, entry := range arg.Entries {
		logIndex := arg.PrevLogIndex + 1 + i
		if logIndex < len(rn.log) {
			if rn.log[logIndex].Term != entry.Term {
				rn.log = rn.log[:logIndex]
				rn.log = append(rn.log, entry)
			}
		} else {
			rn.log = append(rn.log, entry)
		}
	}

	if arg.LeaderCommit > rn.commitIndex {
		lastNewEntry := arg.PrevLogIndex + len(arg.Entries)
		if arg.LeaderCommit < lastNewEntry {
			rn.commitIndex = arg.LeaderCommit
		} else {
			rn.commitIndex = lastNewEntry
		}
	}

	return AppendEntriesReply{Term: rn.currentTerm, Success: true}
}

func getRandomTimeout() time.Duration {
	return time.Duration(150+rand.Intn(150)) * time.Millisecond
}

func (rn *RaftNode) run() {
	rn.mu.Lock()
	rn.log = []LogEntry{{Term: 0, Command: nil}}
	rn.votedFor = nil
	peerCount := len(rn.peers)
	rn.mu.Unlock()

	leaderTime := time.Duration(100) * time.Millisecond
	leaderTimer := time.NewTimer(leaderTime)
	randomTimer := time.NewTimer(getRandomTimeout())

	for {
		rn.mu.Lock()
		currentState := rn.state
		rn.mu.Unlock()

		switch currentState {
		case Follower:
			select {
			case <-randomTimer.C:
				rn.mu.Lock()
				rn.state = Candidate
				rn.mu.Unlock()
				randomTimer.Reset(getRandomTimeout())

			case arg := <-rn.AppendEntriesCh:
				randomTimer.Reset(getRandomTimeout())
				rn.HandleAppendEntries(arg)

			case arg := <-rn.RequestVoteCh:
				randomTimer.Reset(getRandomTimeout())
				rn.HandleRequestVote(arg)

			case req := <-rn.ClientCommandCh:
				req.responseCh <- errors.New("Not Leader")
			}

		case Candidate:
			rn.mu.Lock()
			rn.currentTerm++
			rn.votedFor = &rn.id
			lastLogIndex := len(rn.log) - 1
			lastLogTerm := rn.log[lastLogIndex].Term
			rn.mu.Unlock()

			randomTimer.Reset(getRandomTimeout())

			argRV := RequestVote{
				Term:         rn.currentTerm,
				CandidateId:  rn.id,
				LastLogIndex: lastLogIndex,
				LastLogTerm:  lastLogTerm,
			}

			votes := 1
			replyCh := make(chan RequestVoteReply, peerCount)

			for i := 0; i < peerCount; i++ {
				if i == rn.id {
					continue
				}
				go func(peer int) {
					var reply RequestVoteReply
					if ok := rn.sendRequestVote(peer, argRV, &reply); ok {
						replyCh <- reply
					}
				}(i)
			}

		ElectionLoop:
			for {
				select {
				case reply := <-replyCh:
					if reply.VoteGranted {
						votes++
						if votes > peerCount/2 {
							rn.mu.Lock()
							rn.state = Leader
							rn.nextIndex = make([]int, peerCount)
							rn.matchIndex = make([]int, peerCount)
							for i := range rn.peers {
								rn.nextIndex[i] = len(rn.log)
								rn.matchIndex[i] = 0
							}
							rn.mu.Unlock()
							break ElectionLoop
						}
					}
				case arg := <-rn.AppendEntriesCh:
					rn.HandleAppendEntries(arg)
					break ElectionLoop
				case arg := <-rn.RequestVoteCh:
					rn.HandleRequestVote(arg)
				case req := <-rn.ClientCommandCh:
					req.responseCh <- errors.New("Not Leader")
				case <-randomTimer.C:
					break ElectionLoop
				}
			}

		case Leader:
			select {
			case <-leaderTimer.C:
				for i := 0; i < peerCount; i++ {
					if i == rn.id {
						continue
					}
					rn.mu.Lock()
					prevIdx := rn.nextIndex[i] - 1
					arg := AppendEntries{
						Term:         rn.currentTerm,
						LeaderId:     rn.id,
						PrevLogIndex: prevIdx,
						PrevLogTerm:  rn.log[prevIdx].Term,
						Entries:      []LogEntry{}, // Empty for heartbeat
						LeaderCommit: rn.commitIndex,
					}
					rn.mu.Unlock()

					go func(server int, arg AppendEntries) {
						var reply AppendEntriesReply
						if ok := rn.sendAppendEntries(server, arg, &reply); !ok {
							return
						}

						rn.mu.Lock()
						defer rn.mu.Unlock()
						if reply.Term > rn.currentTerm {
							rn.currentTerm = reply.Term
							rn.state = Follower
							rn.votedFor = nil
						}
					}(i, arg)
				}
				leaderTimer.Reset(leaderTime)

			case req := <-rn.ClientCommandCh:
				rn.mu.Lock()
				rn.log = append(rn.log, LogEntry{
					Command: req.command,
					Term:    rn.currentTerm,
				})
				rn.mu.Unlock()

				req.responseCh <- nil

				for i := 0; i < peerCount; i++ {
					if i == rn.id {
						continue
					}

					rn.mu.Lock()
					if len(rn.log)-1 >= rn.nextIndex[i] {
						rn.mu.Unlock()
						
						go func(server int) {
							var reply AppendEntriesReply
							for {
								rn.mu.Lock()
								if rn.state != Leader {
									rn.mu.Unlock()
									return
								}

								prevIdx := rn.nextIndex[server] - 1
								args := AppendEntries{
									Term:         rn.currentTerm,
									LeaderId:     rn.id,
									PrevLogIndex: prevIdx,
									PrevLogTerm:  rn.log[prevIdx].Term,
									Entries:      rn.log[rn.nextIndex[server]:],
									LeaderCommit: rn.commitIndex,
								}
								rn.mu.Unlock()

								if ok := rn.sendAppendEntries(server, args, &reply); !ok {
									return
								}

								rn.mu.Lock()
								if reply.Term > rn.currentTerm {
									rn.currentTerm = reply.Term
									rn.state = Follower
									rn.votedFor = nil
									rn.mu.Unlock()
									return
								}

								if reply.Success {
									rn.nextIndex[server] = args.PrevLogIndex + len(args.Entries) + 1
									rn.matchIndex[server] = rn.nextIndex[server] - 1
									rn.mu.Unlock()
									return
								} else {
									rn.nextIndex[server]--
									rn.mu.Unlock()
								}
							}
						}(i)
					} else {
						rn.mu.Unlock()
					}
				}

			case arg := <-rn.AppendEntriesCh:
				rn.HandleAppendEntries(arg)
			case arg := <-rn.RequestVoteCh:
				rn.HandleRequestVote(arg)
			}
		}
	}
}

func (rn *RaftNode) ClientCommand(command interface{}) error {
	if command == nil {
		return errors.New("nil command")
	}
	req := clientRequest{
		command:    command,
		responseCh: make(chan error),
	}

	rn.ClientCommandCh <- req
	return <-req.responseCh
}