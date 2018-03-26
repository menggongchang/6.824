package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import "sync"
import "labrpc"
import (
	"math/rand"
	"strconv"
	"time"
)

// import "bytes"
// import "encoding/gob"

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make().
//
type ApplyMsg struct {
	Index       int
	Command     interface{}
	UseSnapshot bool   // ignore for lab2; only used in lab3
	Snapshot    []byte // ignore for lab2; only used in lab3
}

//
// A Go object implementing a single Raft peer.
//
type Log struct {
	Term    int //命令接受时的任期号
	Command interface{}
}

//raft节点
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]

	//持久保存
	currentTerm int
	votedFor    int   //-1：本轮未投票
	logs        []Log //log entries
	//易变的状态,all servers
	commitIndex int
	lastApplied int //应用到状态机
	//易变的状态,only leaders
	nextIndex  []int
	matchIndex []int
	//自己添加的变量
	state        int       //表示server的状态，1follower，2candidate，3leader
	timeOutIndex int       //表示当前选举倒计时是第几轮
	timeOut      chan bool //是否倒计时完成
}

// return currentTerm and whether this server believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var isleader bool
	if rf.state == 3 {
		isleader = true
	} else {
		isleader = false
	}
	return rf.currentTerm, isleader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := gob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := gob.NewDecoder(r)
	// d.Decode(&rf.xxx)
	// d.Decode(&rf.yyy)
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
}

//打印
func MyLog(raftIndex int, state int, timeOutIndex int, message string) {
	var role string
	switch state {
	case 1:
		role = "follower"
	case 2:
		role = "candidate"
	case 3:
		role = "leader"
	}
	ZM.Println("Server:", raftIndex, role, ",我的倒计时轮数 ", timeOutIndex, ",Msg: ", message)
}

//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	switch rf.state {
	case 1:
	case 2:
		isLeader = false
	case 3:
		isLeader = true
		term = rf.currentTerm
		index = len(rf.logs) + 1
		rf.logs = append(rf.logs, Log{term, command})
	}
	return index, term, isLeader
}

//
// the tester calls Kill() when a Raft instance won't
// be needed again. you are not required to do anything
// in Kill(), but it might be convenient to (for example)
// turn off debug output from this instance.
//
func (rf *Raft) Kill() {
	// Your code here, if desired.
	ZM.Println("Raft instance Kill")
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
// raft节点创建过程
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = append(rf.peers, peers...)
	rf.persister = persister
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).

	//状态初始化
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.logs = make([]Log, 0)
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))
	rf.timeOutIndex = 0
	rf.timeOut = make(chan bool)
	rf.state = 1 //follower
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	go work(rf)

	return rf
}

//发起新一轮的倒计时，首先拿到此轮标识
func newElectionTimeOutIndex(rf *Raft) int {
	rf.mu.Lock()
	rf.timeOutIndex++
	index := rf.timeOutIndex
	rf.mu.Unlock()
	return index
}

//发起选举倒计时
func startElectionTimeOut(rf *Raft, index int) {
	MyLog(rf.me, rf.state, index, "开始新一轮倒计时")

	//倒计时随机
	time.Sleep(time.Millisecond * 100)
	randNum := rand.Intn(10)
	for i := 0; i < randNum; i++ {
		time.Sleep(time.Millisecond * 10)
	}

	//倒计时结束，比较当前倒计时是否是自己这一轮; 因为leader心跳包会使自己重启倒计时，这次的就不管了
	rf.mu.Lock()
	indexNew := rf.timeOutIndex
	rf.mu.Unlock()

	//有效的倒计时结束
	if index == indexNew {
		switch rf.state {
		case 1:
			rf.state = 2
			rf.timeOut <- true //选举倒计时结束
			MyLog(rf.me, rf.state, index, "结束倒计时，成为候选人")
		case 2:
			rf.timeOut <- true //选举倒计时结束
			MyLog(rf.me, rf.state, index, "结束倒计时，重新发起选举")
		case 3:
		}

	} else { //新的倒计时已经开始，此轮无效
		ZM.Println("server:", rf.me, rf.state, "倒计时第", index, "轮结束，但是无效，新的已经开始了。", indexNew)
		return
	}
}

//结束当前的倒计时
func endElectionTimeOut(rf *Raft, timeOutIndex int, serverState int) {
	MyLog(rf.me, rf.state, timeOutIndex, "被结束,成为状态"+strconv.Itoa(serverState))

	rf.mu.Lock()
	index := rf.timeOutIndex
	rf.mu.Unlock()

	if index == timeOutIndex {
		rf.state = serverState
		rf.timeOut <- true
	} else {
		return //新的倒计时已经开始，此轮无效
	}
}

//raft节点的循环
func work(rf *Raft) {
	for {
		switch rf.state {
		case 1:
			index := newElectionTimeOutIndex(rf)
			go startElectionTimeOut(rf, index) //启动选举倒计时
			<-rf.timeOut                       //完成选举倒计时
		case 2:
			index := newElectionTimeOutIndex(rf)
			go startElectionTimeOut(rf, index) //启动选举倒计时
			go candidateWork(rf, index)        //发起选举
			<-rf.timeOut                       //完成选举倒计时
		case 3:
			//leader不用倒计时，有问题就变follower
			leaderWork(rf)
		}
	}
}

//候选人发起选举
func candidateWork(rf *Raft, electionIndex int) {
	MyLog(rf.me, rf.state, electionIndex, "发起选举")
	rf.currentTerm++
	rf.votedFor = rf.me
	serverNum := len(rf.peers)

	//发送RequestVote RPC
	lastLogIndex := len(rf.logs)
	lastLogTerm := 0
	if lastLogIndex != 0 {
		lastLogTerm = rf.logs[len(rf.logs)-1].Term
	}
	args := RequestVoteArgs{rf.currentTerm, rf.me, lastLogIndex, lastLogTerm}
	votes := make(chan RequestVoteReply, serverNum-1) //接收到的投票
	for i := 0; i < serverNum && i != rf.me; i++ {
		go func(serverID int) {
			reply := RequestVoteReply{-1, false}
			ok := rf.sendRequestVote(serverID, &args, &reply)
			if ok {
				votes <- reply
			} else {
				return
			}
		}(i)
	}

	//统计投票结果
	voteNum := 1 //自已有一票
	for vote := range votes {
		//停止选举
		if vote.Term > rf.currentTerm {
			rf.currentTerm = vote.Term
			rf.state = 1
			rf.votedFor = -1
			endElectionTimeOut(rf, electionIndex, 1)
			break
		}
		if vote.VoteGranted {
			voteNum++
		}
		if voteNum > (serverNum-1)/2 {
			//选举成功
			endElectionTimeOut(rf, electionIndex, 3)
			break
		}
	}
}

func leaderWork(rf *Raft) {
	ZM.Printf("server %d do leader work.\n", rf.me)
	//leader每次需要重置
	leaderLastLog := len(rf.logs)
	for i := 0; i < len(rf.peers); i++ {
		rf.nextIndex[i] = leaderLastLog + 1
		rf.matchIndex[i] = 0
	}

	for rf.state == 3 {
		ZM.Printf("server %d send AppendEntries\n", rf.me)
		leaderAppendEntries(rf)
		time.Sleep(time.Millisecond * 30)
	}
}

//定期发送心跳包
func leaderAppendEntries(rf *Raft) {
	//心跳包
	args := AppendEntriesArgs{
		rf.currentTerm,
		rf.me,
		0,
		0,
		make([]Log, 0),
		0,
	}
	for i := 0; i < len(rf.peers) && i != rf.me; i++ {
		go func(serverID int) {
			reply := AppendEntriesReply{rf.currentTerm, true}
			ok := rf.sendAppendEntries(serverID, &args, &reply)
			if ok && reply.Term > rf.currentTerm {
				rf.currentTerm = reply.Term
				rf.state = 1
				rf.votedFor = -1
			}
		}(i)
	}
}
