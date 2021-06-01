
The Mission : An EV company is about to begin delivery of a massive number vehicles. They need to communicate with all vehicles by their VIN numbers (string) in real-time to:
* Collect location, assume (x, y) coordinates as integers only for this exercise, on a regular basis (5 seconds intervals)
* Collect current speed (integer)
* Inquire about the drive status ("parked", "driving", "reverse", a string)
* Send commands to the car to "honk", "toggle headlights", "toggle door lock"

Application is built using Go programming language and JSON RPC used as communitcation mechanism

**Build Process:**
- To build just create Go environment (~/go/src/) and get this ev folder in it.
- To build app run below script
  ./build.sh

**Launching Apps: **
  - By default Server runs on localhost on port 8009
  - By default Client(Vehicle) connects with local server on port 8009 

Go to work directory 
   cd ~/go/src/ev

**To launch Server: **

 ./ev server 

**To launch Client(Vehicle Instance): **

 ./ev client <Optional ServerIP:8009>

**To launch CLI: **

 ./ev cli
