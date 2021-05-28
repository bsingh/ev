Here are few comments on each questions :

Build Process: 
- Application is built using Go programming language and JSON RPC used as communitcation mechanism
- To build just create Go environment (~/go/src/) and get this ev folder in it.
- To build app run below script
./build.sh

Launching Apps: 
  - By default Server runs on localhost on port 8009
  - By default Client(Vehicle) connects with local server on port 8009 

Go to work directory 
   cd ~/go/src/ev

For Server: 
 ./ev server 

For Client(Vehicle Instance):

 ./ev client <Optional ServerIP:8009>

For CLI: 
 ./ev cli
