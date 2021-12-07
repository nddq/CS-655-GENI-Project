package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

const (
	allChar     = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	host        = "localhost"
	port        = "80"
	protocol    = "tcp"
	maxPwd      = 380204031 // maximum number that represents the last 5 character password
	maxJobBatch = 380
)

var HashCh chan string

type Master struct {
	mu             sync.Mutex
	curHash        string
	curJobBatch    int
	workLeft       bool
	workQueue      []string
	maxWorker      int
	curNoOfWorker  int
	workerReported int
	startedTime    time.Time
}

type ReportArgs struct {
	PwdFound   bool
	Pwd        string
	JobBatchNo int
}

type ReportReply struct {
	Ack bool
}

type GetWorkArgs struct {
}

type GetWorkReply struct {
	LowerBound, UpperBound int64
	JobBatchNo             int
	WantedHash             string
	WorkLeft               bool
}

func enqueue(queue []string, element string) []string {
	queue = append(queue, element)
	return queue
}

func dequeue(queue []string) (string, []string) {
	element := queue[0]
	return element, queue[1:]
}

func isEmpty(queue []string) bool {
	return len(queue) == 0
}

func (m *Master) NewHashToCrack(hash string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.workLeft {
		log.Printf("Working on a diffent hash. Enqueuing new hash %s\n", hash)
		m.workQueue = enqueue(m.workQueue, hash)
	} else {
		m.startedTime = time.Now()
		m.curHash = hash
		m.curJobBatch = 0
		m.workLeft = true
	}
}

func (m *Master) Report(args *ReportArgs, reply *ReportReply) error {
	log.Printf("Job batch number %d reported\n", args.JobBatchNo)
	m.mu.Lock()
	defer m.mu.Unlock()
	if args.PwdFound {
		log.Printf("Password found in job batch %d : %s\n", args.JobBatchNo, args.Pwd)
		elapsed := time.Since(m.startedTime)
		m.workLeft = false
		log.Printf("Time elapsed: %s\n", elapsed)
		reply.Ack = true

		if !isEmpty(m.workQueue) {
			var nextHash string
			nextHash, m.workQueue = dequeue(m.workQueue)
			log.Printf("Working on a next hash %s\n", nextHash)
			m.startedTime = time.Now()
			m.curHash = nextHash
			m.curJobBatch = 0
			m.workLeft = true
		}
	} else {
		if args.JobBatchNo == maxJobBatch {
			log.Printf("Unable to crack hash\n")
			if !isEmpty(m.workQueue) {
				var nextHash string
				nextHash, m.workQueue = dequeue(m.workQueue)
				log.Printf("Working on a next hash %s\n", nextHash)
				m.startedTime = time.Now()
				m.curHash = nextHash
				m.curJobBatch = 0
				m.workLeft = true
			}
		}
		reply.Ack = true
	}
	return nil
}

func (m *Master) GetWork(args *GetWorkArgs, reply *GetWorkReply) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.workLeft {
		reply.WorkLeft = false
		return nil
	}
	reply.LowerBound = int64(1000000 * m.curJobBatch)
	reply.UpperBound = reply.LowerBound + (1000000 - 1)
	reply.JobBatchNo = m.curJobBatch
	reply.WorkLeft = true
	reply.WantedHash = m.curHash
	if reply.UpperBound > maxPwd {
		reply.UpperBound = maxPwd
		m.workLeft = false
	}
	m.curJobBatch++
	return nil
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Password Cracker API v1.0")
	log.Printf("Endpoint Hit: homePage\n")
}

func getPwd(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	log.Printf("Endpoint Hit: getPwd\n")
	HashCh <- hash
}

func handleAPIRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/getPwd/{hash}", getPwd)

	log.Printf("Starting API server on %s:%s", host, "10000")

	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func (m *Master) run() {
	for {
		select {
		case hash := <-HashCh:
			log.Printf("Received hash to crack : %s", hash)
			m.NewHashToCrack(hash)
		default:
			continue
		}

	}
}

func main() {
	master := new(Master)
	master.curJobBatch = 0
	master.workLeft = false
	HashCh = make(chan string)
	rpc.Register(master)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	go master.run()

	go handleAPIRequests()

	tcpAddr, err := net.ResolveTCPAddr(protocol, (host + ":" + port))
	checkError(err)
	listener, err := net.ListenTCP(protocol, tcpAddr)
	checkError(err)

	log.Printf("Starting master's server on %s:%s", host, port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		log.Printf("Worker %s connected.\n", conn.RemoteAddr().String())
		go rpc.ServeConn(conn)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
