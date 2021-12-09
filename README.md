# Distributed Password Cracker
### Project member
- Quang Nguyen
### Links
- Github : https://github.com/nddq/CS-655-GENI-Project
- GENI : https://portal.geni.net/secure/slice.php?slice_id=434bc9ae-9887-4954-8797-1263e5872f7c
### Introduction
- This project aims to implement a distributed passwork cracker using Go and running on GENI
- This project shows how weak MD5 hashes are given enough computing power
- Assumtions:
-- Both the coordinator and the worker never crash.
-- No network packet loss

### Running instruction
- Reserve resources on GENI using the given RSpec file on Github
- SSH into the coordinator node, wget the shell script for the coordinator node by running ``` sudo wget https://raw.githubusercontent.com/nddq/CS-655-GENI-Project/main/coordinator.sh```. Run the script using ```sh coordinator.sh```
- After the script runs, an API server and the coordinator should be up and running, if not, simply ```cd``` into the coordinator within the project folder and run ```go run main.go```
- SSH into the workers node, wget the shell script for the worker node by running ```sudo wget https://raw.githubusercontent.com/nddq/CS-655-GENI-Project/main/worker.sh```. Run the script using ```sh worker.sh```
- ```cd``` into the worker folder and run ```go run main.go 10.10.<worker number>.11``` to connect to the coordinator server. After a worker node is connected, a log should be output to the coordinator's terminal.
### Experimental method
![Image of Architecture](https://i.imgur.com/oM8A1OH.png)
- The program consists of a coordinator node and multiple worker nodes running on different GENI machine
- The coordinator node runs both the coordinator server, which is used by the workers to get jobs, and the API server, which users can make hash cracking requests.
- A request queue is implemented at the coordinator node, meaning that multiple requests can be make to the node.
- After a worker connected to the coordinator, it will make a polling request every 10 seconds for job.
- User submits a hash of a 5 character password consists of lower and upper case letters (A-Z, a-z). After the coordinator received a hash through the API, it immediately initialize the job batch number to 0, which will be use by the workers to figure out which range of integer to work on; each integer in this range will be convert to their equivalent 5 characters password. There are 1 million integer for each job batch, meaning that each worker will try to hash 1 million passwords to find the hash that the coordinator received. Once a password is found, the coordinator will be notify and return the result to the user, as well as the elapsed time since the coordinator received the request. If all job batches have been go through and no password is found, the coordinator will send this result to the user as well.
- The workers used RPC (Remote Procedure Call), a protocol runs on top of TCP, to communicate with the coordinator.
- Due to the coordinator being a passive listener, the program can be easily scale up by starting new worker nodes and connect them to the coordinator; the coordinator will hand them work if there is any.
- Different number of workers can be use to test how fast the system is able to crack a given hash
- The coordinator utilizes Go routines to handle workers' connection, which can support thousands of connections concurrently, given enough computing power.

### Results

##### Usage instruction
- Start coordinator server and connect workers
![Image of Architecture](https://i.imgur.com/Xc25AAw.png)
- Submit work through API, coordinator starts handing out work when worker request.
![Image of Architecture](https://i.imgur.com/DUhzaYN.png)
- Password found, coordinator output password and time taken to crack the hash.
![Image of Architecture](https://i.imgur.com/uMhQxeR.png)

##### Analysis
- Graph that shows the time it takes to crack the hash of password 'zzzzz' against the number of workers used.
![Image of Architecture](https://i.imgur.com/XG43G97.png)
- It's interesting to see as we add more worker nodes, the time it takes to crack the hash decrease exponentially.

### Conclusion
- The program is able to crack more than 300 millions 5 character passwords in a short amount of time. Given sufficient computing power and more worker nodes, any MD5 hash can be crack within a reasonable amount of time, showing how weak it is and why most systems nowaday don't use it.
- Some extensions that I could implement in the future include add functions to simulate network delay, network drop, as well as worker's node failing in order to test the system's fault torelant. A way to scale down the number of workers used would be beneficial. A web UI would be helpful for users to interact with the system.

### Division of labor
- Quang Nguyen: Everything

