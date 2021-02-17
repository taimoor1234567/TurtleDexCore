module github.com/turtledex/TurtleDexCore

go 1.13

require (
	github.com/aead/chacha20 v0.0.0-20180709150244-8b13a72661da
	github.com/dchest/threefish v0.0.0-20120919164726-3ecf4c494abf
	github.com/hanwen/go-fuse/v2 v2.0.2
	github.com/inconshreveable/go-update v0.0.0-20160112193335-8152e7eb6ccf
	github.com/julienschmidt/httprouter v1.3.0
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/klauspost/cpuid v1.2.2 // indirect
	github.com/klauspost/reedsolomon v1.9.3
	github.com/montanaflynn/stats v0.6.3
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/vbauerster/mpb/v5 v5.0.3
	github.com/xtaci/smux v1.3.3
	github.com/turtledex/bolt v1.4.4
	github.com/turtledex/demotemutex v0.0.0-20151003192217-235395f71c40
	github.com/turtledex/encoding v0.0.0-20200604091946-456c3dc907fe
	github.com/turtledex/entropy-mnemonics v0.0.0-20181018051301-7532f67e3500
	github.com/turtledex/errors v0.0.0-20200929122200-06c536cf6975
	github.com/turtledex/fastrand v0.0.0-20181126182046-603482d69e40
	github.com/turtledex/go-upnp v0.0.0-20181011194642-3a71999ed0d3
	github.com/turtledex/log v0.0.0-20200604091839-0ba4a941cdc2
	github.com/turtledex/merkletree v0.0.0-20200118113624-07fbf710afc4
	github.com/turtledex/monitor v0.0.0-20191205095550-2b0fd3e1012a
	github.com/turtledex/ratelimit v0.0.0-20200811080431-99b8f0768b2e
	github.com/turtledex/siamux v0.0.0-20210210103854-9bdf3025036b
	github.com/turtledex/threadgroup v0.0.0-20200608151952-38921fbef213
	github.com/turtledex/writeaheadlog v0.0.0-20200618142844-c59a90f49130
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae // indirect
	golang.org/x/text v0.3.3 // indirect
)

replace github.com/xtaci/smux => ./vendor/github.com/xtaci/smux
