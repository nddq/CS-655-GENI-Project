package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"time"
)

const (
	allChar  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	port     = "8080"
	protocol = "tcp"
)

func convertToPwd(num int64) string {
	s := ""
	if num == 0 {
		s = string(allChar[num])
	} else {
		for num > 0 {
			if num < 52 {
				s = string(allChar[num]) + s
				num = 0
			} else {
				rem := num % 52
				s = string(allChar[rem]) + s
				num = (num - rem) / 52
			}
		}
	}
	if len(s) < 5 {
		tmp := len(s)
		for i := 0; i < 5-tmp; i++ {
			s = "A" + s
		}
	}
	return s
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

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Missing args\n")
		os.Exit(1)
	}
	host := os.Args[1]

	client, err := rpc.Dial(protocol, (host + ":" + port))
	if err != nil {
		log.Fatal("dialing:", err)
	}
	for {
		args := GetWorkArgs{}
		reply := GetWorkReply{}
		err = client.Call("Coordinator.GetWork", args, &reply)
		if err != nil {
			log.Fatal("GetWork error:", err)
		}
		for !reply.WorkLeft {
			fmt.Printf("No work left\n")
			time.Sleep(10 * time.Second)
			fmt.Printf("Attempting to get work...\n")
			err = client.Call("Coordinator.GetWork", args, &reply)
		}
		fmt.Printf("Got work with lower bound %s and upper bound %s\n", convertToPwd(reply.LowerBound), convertToPwd(reply.UpperBound))
		pwdFound := false
		actualPwd := ""
		for i := reply.LowerBound; i <= reply.UpperBound; i++ {
			pwd := convertToPwd(i)
			hash := fmt.Sprintf("%x", md5.Sum([]byte(pwd)))
			if hash == reply.WantedHash {
				fmt.Printf("Hashes matched. Password found : %s\n", pwd)
				actualPwd = pwd
				pwdFound = true
				break
			}
		}
		reportArgs := ReportArgs{}
		reportReply := ReportReply{}
		reportArgs.PwdFound = pwdFound
		reportArgs.Pwd = actualPwd
		reportArgs.JobBatchNo = reply.JobBatchNo
		err = client.Call("Coordinator.Report", reportArgs, &reportReply)
		if err != nil || !reportReply.Ack {
			log.Fatal("Report error:", err)
		}

	}
}
