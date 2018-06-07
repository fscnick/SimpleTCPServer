package main

type Options struct {
	// Bool
	debug				bool
	version      		bool
	help				bool

	// Int
	maxgo				int
	worker				int
	incomingTimeout		int
	outgoingTimeout		int
	bufferSize			int

	// String
	listen				string
	to 					string
	httpListenAddr		string
	statusPath			string
}