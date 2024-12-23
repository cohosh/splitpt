# SplitPT

SplitPT implements traffic splitting for existing PTs supported by Tor.

SplitPT itself is a pluggble transport that acts as a shim between the client, hops, and server.

## Client Usage

TODO

## Server Usage

TODO

## To Run

You'll need three separate terminal windows, called A, B, and C for the purposes of this README.

1. In terminal A, launch the splitpt server
2. In terminal B, launch the obfs4 server
3. In terminal C, launch the splitpt client
