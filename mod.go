package dvoting

import "github.com/prometheus/client_golang/prometheus"

// Version contains the current or build version. This variable can be changed
// at build time with:
//
//	go build -ldflags="-X 'github.com/dedis/d-voting.Version=v1.0.0'"
//
// Version should be fetched from git: `git describe --tags`
var Version = "unknown"

// BuildTime indicates the time at which the binary has been built. Must be set
// as with Version.
var BuildTime = "unknown"

// PromCollectors exposes the Prometheus collectors created in d-Voting.
var PromCollectors []prometheus.Collector

var promBuildInfos = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "dvoting_build_info",
	Help: "build info about the d-voting system",
	ConstLabels: prometheus.Labels{
		"version":    Version,
		"build_time": BuildTime,
	},
})

func init() {
	PromCollectors = append(PromCollectors, promBuildInfos)
	promBuildInfos.Set(1)
}
