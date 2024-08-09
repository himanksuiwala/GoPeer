
## 1.When giving some addresses which does not exists:
server := instantiatePeer(":3000")
server2 := instantiatePeer(":4000", ":3100")
server3 := instantiatePeer(":5000", ":14000", ":3000")
server4 := instantiatePeer(":6000", ":5000", ":4000", ":31000")

### OUTPUT:
Spinning up the peer :3000
[Server@3000]: Server is up & listening...
Spinning up the peer :4000
[Server@4000]: Server is up & listening...
[Server@4000]: Trying to connect with peers... :3100
[Server@4000]: Unable to connect with Peer :3100, peer is unavailable for connection..
Spinning up the peer :5000
[Server@5000]: Server is up & listening...
[Server@5000]: Trying to connect with peers... :14000, :3000
Spinning up the peer :3000
[Server@3000]: Server is up & listening...
Spinning up the peer :4000
[Server@4000]: Server is up & listening...
[Server@4000]: Trying to connect with peers... :3100
[Server@4000]: Unable to connect with Peer :3100, peer is unavailable for connection..
Spinning up the peer :5000
[Server@5000]: Server is up & listening...
[Server@5000]: Trying to connect with peers... :14000, :3000
[Server@4000]: Unable to connect with Peer :3100, peer is unavailable for connection..
Spinning up the peer :5000
[Server@5000]: Server is up & listening...
[Server@5000]: Trying to connect with peers... :14000, :3000
[Server@5000]: Trying to connect with peers... :14000, :3000
[Server@5000]: Unable to connect with Peer :14000, peer is unavailable for connection..
Spinning up the peer :6000
[Server@3000]: New peer.0.1:57511 connected
[Server@6000]: Server is up & listening...
[Server@6000]: Trying to connect with peers... :5000, :4000, :31000
[Server@5000]: Connected with Peer:3000
[Server@6000]: Connected with Peer:5000
[Server@5000]: New peer.0.1:57512 connected
[Server@6000]: Connected with Peer:4000
[Server@4000]: New peer.0.1:57513 connected
[Server@6000]: Unable to connect with Peer :31000, peer is unavailable for connection..

## 2.When giving some addresses which does exists:

### OUTPUT:
Spinning up the peer :3000
[Server@3000]: Server is up & listening...
Spinning up the peer :4000
[Server@4000]: Server is up & listening...
[Server@4000]: Trying to connect with peers... :3000
Spinning up the peer :3000
[Server@3000]: Server is up & listening...
Spinning up the peer :4000
[Server@4000]: Server is up & listening...
[Server@4000]: Trying to connect with peers... :3000
Spinning up the peer :4000
[Server@4000]: Server is up & listening...
[Server@4000]: Trying to connect with peers... :3000
[Server@4000]: Trying to connect with peers... :3000
Spinning up the peer :5000
[Server@4000]: Connected with Peer:3000
[Server@3000]: New peer.0.1:58036 connected
[Server@5000]: Server is up & listening...
[Server@5000]: Trying to connect with peers... :4000, :3000
[Server@5000]: Connected with Peer:4000
[Server@4000]: New peer.0.1:58037 connected
Spinning up the peer :6000
[Server@5000]: Connected with Peer:3000
[Server@3000]: New peer.0.1:58038 connected
[Server@6000]: Server is up & listening...
[Server@6000]: Trying to connect with peers... :5000, :4000, :3000
[Server@5000]: New peer.0.1:58039 connected
[Server@6000]: Connected with Peer:5000
[Server@4000]: New peer.0.1:58040 connected
[Server@6000]: Connected with Peer:4000
[Server@6000]: Connected with Peer:3000
[Server@3000]: New peer.0.1:58041 connected
----
Spinning up the peer :3000
[Server@3000]: Storage location validated
[Server@3000]: Server is up & listening...
[Server@3000]: Trying to connect with peers, if any...
Spinning up the peer :4000
[Server@4000]: Storage location validated
[Server@4000]: Server is up & listening...
[Server@4000]: Trying to connect with peers, if any... :3000
[Server@4000]: Connected with Peer:3000
[Server@4000]: Connection Pool...127.0.0.1:3000,
[Server@3000]: New peer.0.1:63333 connected
Spinning up the peer :5000
[Server@5000]: Storage location validated
[Server@5000]: Server is up & listening...
[Server@5000]: Trying to connect with peers, if any... :4000, :3000
[Server@4000]: New peer.0.1:63334 connected
[Server@5000]: Connected with Peer:4000
[Server@5000]: Connection Pool...127.0.0.1:4000,
[Server@5000]: Connected with Peer:3000
[Server@5000]: Connection Pool...127.0.0.1:4000, 127.0.0.1:3000,
[Server@3000]: New peer.0.1:63335 connected
[Server@5000]: Writing 32 bytes to connected peers...
[Peer@:4000]: File incoming from: 127.0.0.1:4000
[Peer@:3000]: File incoming from: 127.0.0.1:3000
[Peer@:3000]: Succesfully written in storage: This data is sent over the network.
[Peer@:4000]: Succesfully written in storage: This data is sent over the network.
----