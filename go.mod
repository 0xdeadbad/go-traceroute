module github.com/ubermenzchen/go-traceroute

go 1.18

require (
	internal/platform v1.0.0
	pkg/traceroute v1.0.0
)

require (
	golang.org/x/net v0.0.0-20220425223048-2871e0cb64e4 // indirect
	golang.org/x/sys v0.0.0-20220503163025-988cb79eb6c6 // indirect
)

replace internal/platform => ./internal/platform

replace pkg/traceroute => ./pkg/traceroute
