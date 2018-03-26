package raft

type RequestVoteArgs struct {
	Term         int //候选人任期号
	CandidateId  int //候选人Id
	LastLogIndex int //
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int  //currentTerm,用来使候选人更新自己；
	VoteGranted bool //候选人是否赢得了投票
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	ZM.Printf("Server %d accept RequestVote from %d\n ", rf.me, args.CandidateId)
	//比较任期号

	switch {
	case args.Term < rf.currentTerm:
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
	case args.Term == rf.currentTerm:
		reply.VoteGranted = false
	case args.Term > rf.currentTerm:
		rf.currentTerm = args.Term
		if rf.state == 2 { //candidate->follower
			rf.state = 1
			rf.votedFor = -1
			if (rf.votedFor == -1 || rf.votedFor == args.CandidateId) && NewLogVote(args, rf) {
				reply.VoteGranted = true
				rf.votedFor = args.CandidateId
			} else {
				reply.VoteGranted = false
			}
			ZM.Printf("收到server： %d请求投票，server %d 停止本次选举倒计时\n", args.CandidateId, rf.me)
			endElectionTimeOut(rf, rf.timeOutIndex, 1)
		}
		if rf.state == 3 { //leader->follower
			rf.state = 1
			rf.votedFor = -1
			if (rf.votedFor == -1 || rf.votedFor == args.CandidateId) && NewLogVote(args, rf) {
				reply.VoteGranted = true
				rf.votedFor = args.CandidateId
			} else {
				reply.VoteGranted = false
			}
		}
	}
}

//判断请求投票者是否比自己的日志更新
func NewLogVote(args *RequestVoteArgs, rf *Raft) bool {
	myLastLogIndex := len(rf.logs)
	myLastLogTerm := 0
	if myLastLogIndex != 0 {
		myLastLogTerm = rf.logs[myLastLogIndex-1].Term
	}

	if args.LastLogTerm < myLastLogTerm {
		return false
	} else if args.LastLogTerm > myLastLogTerm {
		return true
	} else if args.LastLogIndex < myLastLogIndex {
		return false
	} else {
		return true
	}
}

// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

//心跳包
type AppendEntriesArgs struct {
	Term         int //leader任期号
	LeaderId     int //
	PrevLogIndex int //
	PrevLogTerm  int
	Entries      []Log //心跳包时为空
	LeaderCommit int   //leader commit index
}
type AppendEntriesReply struct {
	Term    int  //currentTerm ,for leader update
	Success bool //log entry
}

//leader 发送心跳包
func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

//处理心跳包Your code here (2A, 2B).
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	ZM.Printf("Server %d accept AppendEntries from %d ", rf.me, args.LeaderId)

	//比较任期号
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		reply.Success = false
	}

	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		reply.Success = true
		switch rf.state {
		case 1:
			endElectionTimeOut(rf, rf.timeOutIndex, 1)
		case 2:
			rf.state = 1
			rf.votedFor = -1
			endElectionTimeOut(rf, rf.timeOutIndex, 1)
		case 3:
			rf.state = 1
			rf.votedFor = -1
		}
		return
	}

	if args.Term == rf.currentTerm {
		reply.Success = true
		switch rf.state {
		case 1:
			endElectionTimeOut(rf, rf.timeOutIndex, 1)
		case 2:
			rf.state = 1
			endElectionTimeOut(rf, rf.timeOutIndex, 1)
		}
		return
	}

}
