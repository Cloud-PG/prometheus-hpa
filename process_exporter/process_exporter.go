package main

// Author: Valentin Kuznetsov <vkuznet [AT] gmail {DOT} com>
// Example of cmsweb data-service exporter for prometheus.io

import (
	"flag"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/procfs"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

var (
	listeningAddress = flag.String("address", ":18000", "address to expose metrics on web interface.")
	metricsEndpoint  = flag.String("endpoint", "/metrics", "Path under which to expose metrics.")
	scrapeURI        = flag.String("uri", "", "URI of server status page we're going to scrape")
	namespace        = flag.String("prefix", "process_exporter", "namespace/prefix to use")
	pid              = flag.Int("pid", 0, "PID of the process we're going to scrape")
	verbose          = flag.Bool("verbose", false, "verbose output")
)

type Exporter struct {
	URI   string
	mutex sync.Mutex

	// metrics from process collector
	cpuTotal        *prometheus.Desc
	openFDs, maxFDs *prometheus.Desc
	vsize, maxVsize *prometheus.Desc
	rss             *prometheus.Desc

	// node specific metrics
	memPercent  *prometheus.Desc
	memTotal    *prometheus.Desc
	memFree     *prometheus.Desc
	swapPercent *prometheus.Desc
	swapTotal   *prometheus.Desc
	swapFree    *prometheus.Desc
	cpuPercent  *prometheus.Desc
	numThreads  *prometheus.Desc
	numCpus     *prometheus.Desc
	load1       *prometheus.Desc
	load5       *prometheus.Desc
	load15      *prometheus.Desc

	//process specific metrics
	procCpu   *prometheus.Desc
	procMem   *prometheus.Desc
	openFiles *prometheus.Desc
	totCon    *prometheus.Desc
	lisCon    *prometheus.Desc
	estCon    *prometheus.Desc
	closeCon  *prometheus.Desc
	timeCon   *prometheus.Desc
}

func NewExporter(uri string) *Exporter {
	return &Exporter{
		URI: uri,
		// metrics from process collector
		cpuTotal: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "process_cpu_seconds_total"),
			"Total user and system CPU time spent in seconds (process collector)",
			nil, nil,
		),
		openFDs: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "process_open_fds"),
			"Number of open file descriptors (process collector)",
			nil, nil,
		),
		maxFDs: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "process_max_fds"),
			"Maximum number of open file descriptors (process collector)",
			nil, nil,
		),
		vsize: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "process_virtual_memory_bytes"),
			"Virtual memory size in bytes (process collector)",
			nil, nil,
		),
		maxVsize: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "process_virtual_memory_max_bytes"),
			"Maximum amount of virtual memory available in bytes (process collector)",
			nil, nil,
		),
		rss: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "process_resident_memory_bytes"),
			"Resident memory size in bytes (process collector)",
			nil, nil,
		),

		// custom metrics
		memPercent: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "memory_percent"),
			"Virtual memory usage of the server",
			nil,
			nil),
		memTotal: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "memory_total"),
			"Virtual total memory usage of the server",
			nil,
			nil),
		memFree: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "memory_free"),
			"Virtual free memory usage of the server",
			nil,
			nil),
		swapPercent: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "swap_percent"),
			"Swap memory usage of the server",
			nil,
			nil),
		swapTotal: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "swap_total"),
			"Virtual total swap usage of the server",
			nil,
			nil),
		swapFree: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "swap_free"),
			"Virtual free swap usage of the server",
			nil,
			nil),
		cpuPercent: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "cpu_percent"),
			"cpu percent of the server",
			nil,
			nil),
		numThreads: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "num_threads"),
			"Number of threads",
			nil,
			nil),
		numCpus: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "num_cpus"),
			"Number of CPUs usable by the current process",
			nil,
			nil),
		load1: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "load1"),
			"Load average in last 1m",
			nil,
			nil),
		load5: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "load5"),
			"Load average in last 5m",
			nil,
			nil),
		load15: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "load15"),
			"Load average in last 15m",
			nil,
			nil),
		procCpu: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "proc_cpu"),
			"process CPU",
			nil,
			nil),
		procMem: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "proc_mem"),
			"process memory",
			nil,
			nil),
		openFiles: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "open_files"),
			"Number of open files",
			nil,
			nil),
		totCon: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "total_connections"),
			"Server TOTAL number of connections",
			nil,
			nil),
		lisCon: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "listen_connections"),
			"Server LISTEN number of connections",
			nil,
			nil),
		estCon: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "established_connections"),
			"Server ESTABLISHED number of connections",
			nil,
			nil),
		closeCon: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "close_wait_connections"),
			"Server CLOSE_WAIT number of connections",
			nil,
			nil),
		timeCon: prometheus.NewDesc(
			prometheus.BuildFQName(*namespace, "", "time_wait_connections"),
			"Server TIME_WAIT number of connections",
			nil,
			nil),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	// metrics from process collector
	ch <- e.cpuTotal
	ch <- e.openFDs
	ch <- e.maxFDs
	ch <- e.vsize
	ch <- e.maxVsize
	ch <- e.rss
	// node specific metrics
	ch <- e.memPercent
	ch <- e.memTotal
	ch <- e.memFree
	ch <- e.swapPercent
	ch <- e.swapTotal
	ch <- e.swapFree
	ch <- e.numThreads
	ch <- e.cpuPercent
	ch <- e.numCpus
	ch <- e.load1
	ch <- e.load5
	ch <- e.load15
	// process specific metrics
	ch <- e.procCpu
	ch <- e.procMem
	ch <- e.totCon
	ch <- e.openFiles
	ch <- e.totCon
	ch <- e.lisCon
	ch <- e.estCon
}

// Collect performs metrics collectio of exporter attributes
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	if err := e.collect(ch); err != nil {
		log.Errorf("Error scraping: %s", err)
	}
	return
}

// helper function which collects exporter attributes
func (e *Exporter) collect(ch chan<- prometheus.Metric) error {
	var mempct, memtot, memfree float64
	if v, e := mem.VirtualMemory(); e == nil {
		mempct = v.UsedPercent
		memtot = float64(v.Total)
		memfree = float64(v.Free)
	}
	var swappct, swaptot, swapfree float64
	if v, e := mem.SwapMemory(); e == nil {
		swappct = v.UsedPercent
		swaptot = float64(v.Total)
		swapfree = float64(v.Free)
	}
	var cpupct float64
	if c, e := cpu.Percent(time.Millisecond, false); e == nil {
		cpupct = c[0] // one value since we didn't ask per cpu
	}
	var load1, load5, load15 float64
	if l, e := load.Avg(); e == nil {
		load1 = l.Load1
		load5 = l.Load5
		load15 = l.Load15
	}

	var cpuTotal, vsize, rss, openFDs, maxFDs, maxVsize float64
	if proc, err := procfs.NewProc(int(*pid)); err == nil {
		if stat, err := proc.NewStat(); err == nil {
			cpuTotal = float64(stat.CPUTime())
			vsize = float64(stat.VirtualMemory())
			rss = float64(stat.ResidentMemory())
		}
		if fds, err := proc.FileDescriptorsLen(); err == nil {
			openFDs = float64(fds)
		}
		if limits, err := proc.NewLimits(); err == nil {
			maxFDs = float64(limits.OpenFiles)
			maxVsize = float64(limits.AddressSpace)
		}
	}

	var procCpu, procMem float64
	var estCon, lisCon, othCon, totCon, closeCon, timeCon, openFiles float64
	var nThreads float64
	if proc, err := process.NewProcess(int32(*pid)); err == nil {
		if v, e := proc.CPUPercent(); e == nil {
			procCpu = float64(v)
		}
		if v, e := proc.MemoryPercent(); e == nil {
			procMem = float64(v)
		}

		if v, e := proc.NumThreads(); e == nil {
			nThreads = float64(v)
		}
		if connections, e := proc.Connections(); e == nil {
			for _, v := range connections {
				if v.Status == "LISTEN" {
					lisCon += 1
				} else if v.Status == "ESTABLISHED" {
					estCon += 1
				} else if v.Status == "TIME_WAIT" {
					timeCon += 1
				} else if v.Status == "CLOSE_WAIT" {
					closeCon += 1
				} else {
					othCon += 1
				}
			}
			totCon = lisCon + estCon + timeCon + closeCon + othCon
		}
		if oFiles, e := proc.OpenFiles(); e == nil {
			openFiles = float64(len(oFiles))
		}
	}

	// metrics from process collector
	ch <- prometheus.MustNewConstMetric(e.cpuTotal, prometheus.CounterValue, cpuTotal)
	ch <- prometheus.MustNewConstMetric(e.openFDs, prometheus.CounterValue, openFDs)
	ch <- prometheus.MustNewConstMetric(e.maxFDs, prometheus.CounterValue, maxFDs)
	ch <- prometheus.MustNewConstMetric(e.vsize, prometheus.CounterValue, vsize)
	ch <- prometheus.MustNewConstMetric(e.maxVsize, prometheus.CounterValue, maxVsize)
	ch <- prometheus.MustNewConstMetric(e.rss, prometheus.CounterValue, rss)
	// node specific metrics
	ch <- prometheus.MustNewConstMetric(e.memPercent, prometheus.CounterValue, mempct)
	ch <- prometheus.MustNewConstMetric(e.memTotal, prometheus.CounterValue, memtot)
	ch <- prometheus.MustNewConstMetric(e.memFree, prometheus.CounterValue, memfree)
	ch <- prometheus.MustNewConstMetric(e.swapPercent, prometheus.CounterValue, swappct)
	ch <- prometheus.MustNewConstMetric(e.swapTotal, prometheus.CounterValue, swaptot)
	ch <- prometheus.MustNewConstMetric(e.swapFree, prometheus.CounterValue, swapfree)
	ch <- prometheus.MustNewConstMetric(e.numCpus, prometheus.CounterValue, float64(runtime.NumCPU()))
	ch <- prometheus.MustNewConstMetric(e.load1, prometheus.CounterValue, load1)
	ch <- prometheus.MustNewConstMetric(e.load5, prometheus.CounterValue, load5)
	ch <- prometheus.MustNewConstMetric(e.load15, prometheus.CounterValue, load15)
	// process specific metrics
	ch <- prometheus.MustNewConstMetric(e.procCpu, prometheus.CounterValue, procCpu)
	ch <- prometheus.MustNewConstMetric(e.procMem, prometheus.CounterValue, procMem)
	ch <- prometheus.MustNewConstMetric(e.numThreads, prometheus.CounterValue, nThreads)
	ch <- prometheus.MustNewConstMetric(e.cpuPercent, prometheus.CounterValue, cpupct)
	ch <- prometheus.MustNewConstMetric(e.openFiles, prometheus.CounterValue, openFiles)
	ch <- prometheus.MustNewConstMetric(e.totCon, prometheus.CounterValue, totCon)
	ch <- prometheus.MustNewConstMetric(e.lisCon, prometheus.CounterValue, lisCon)
	ch <- prometheus.MustNewConstMetric(e.estCon, prometheus.CounterValue, estCon)
	ch <- prometheus.MustNewConstMetric(e.closeCon, prometheus.CounterValue, closeCon)
	ch <- prometheus.MustNewConstMetric(e.timeCon, prometheus.CounterValue, timeCon)
	return nil
}

// main function
func main() {
	flag.Parse()
	exporter := NewExporter(*scrapeURI)
	prometheus.MustRegister(exporter)

	log.Infof("Starting Server: %s", *listeningAddress)
	http.Handle(*metricsEndpoint, promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listeningAddress, nil))
}
