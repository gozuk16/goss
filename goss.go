package goss

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	psnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"

	"github.com/inhies/go-bytesize"
)

const timeformat = "2006/01/02 15:04:05"

func Info() []byte {
	i, _ := host.Info()

	// convert to JSON. String() is also implemented
	//fmt.Println(i)

	n, _ := psnet.Interfaces()
	//fmt.Println(n)
	var ip []string
	for _, v := range n {
		if len(v.Addrs) > 0 {
			for _, a := range v.Addrs {
				ipaddr, ipnet, err := net.ParseCIDR(a.Addr)
				if err != nil {
					fmt.Println(err.Error)
				}
				if ipnet.IP.To4() != nil && !ipnet.IP.IsLoopback() && !ipnet.IP.IsLinkLocalUnicast() {
					//fmt.Println(ipaddr.String())
					ip = append(ip, ipaddr.String())
				}
			}
			//fmt.Println(v.Addrs)
		}
	}

	t, _ := host.SensorsTemperatures()
	//fmt.Println(t)
	var cpu_temp string
	for _, v2 := range t {
		//if v2.Temperature > 0 {
		//	fmt.Println(v2.SensorKey + ": " + strconv.FormatFloat(v2.Temperature, 'f', -1, 64))
		//}
		if v2.SensorKey == "TC0P" {
			cpu_temp = strconv.FormatFloat(v2.Temperature, 'f', -1, 64) + "℃"
			break
		}
	}

	var info map[string]interface{}
	info = map[string]interface{}{
		"hostname":        i.Hostname,
		"os":              i.OS,
		"platform":        i.Platform,
		"platformFamily":  i.PlatformFamily,
		"platformVersion": i.PlatformVersion,
		"kernelArch":      i.KernelArch,
		"uptime":          uptime2string(i.Uptime),
		"bootTime":        time.Unix(int64(i.BootTime), 0).Format(timeformat),
		"serverTime":      time.Now().Format(timeformat),
		"cpuTemperature":  cpu_temp,
		"ipaddr":          ip,
	}
	j, _ := json.Marshal(info)

	return j
}

// uptime2string uptime(経過秒)をuptimeと同じ"0 days, 00:00"形式に変換する
func uptime2string(uptime uint64) string {
	const oneDay int = 60 * 60 * 24

	if int(uptime) > oneDay {
		day := int(uptime) / oneDay
		secondsOfTheDay := day * oneDay
		d := time.Duration(int(uptime)-secondsOfTheDay) * time.Second
		d = d.Round(time.Minute)
		h := d / time.Hour
		d -= h * time.Hour
		m := d / time.Minute
		return fmt.Sprintf("%d days, %d:%02d", day, h, m)
	} else {
		d := time.Duration(int(uptime)) * time.Second
		d = d.Round(time.Minute)
		h := d / time.Hour
		d -= h * time.Hour
		m := d / time.Minute
		return fmt.Sprintf("%d:%02d", h, m)
	}
}

func Cpu() []byte {
	c, _ := cpu.Percent(time.Millisecond*200, false)
	core, _ := cpu.Percent(time.Millisecond*200, true)
	//fmt.Printf("%f%%\n", c)

	var cpu map[string]interface{}
	total := uint(c[0])
	var p = []uint{}
	for _, v := range core {
		p = append(p, uint(v))
	}
	cpu = map[string]interface{}{
		"total":  total,
		"percpu": p,
	}
	j, _ := json.Marshal(cpu)

	return j
}

func Load() []byte {
	l, _ := load.Avg()
	//fmt.Printf("%f\n", l)

	var loadAvg map[string]interface{}
	loadAvg = map[string]interface{}{
		"load1":  strconv.FormatFloat(l.Load1, 'f', 2, 64),
		"load5":  strconv.FormatFloat(l.Load5, 'f', 2, 64),
		"load15": strconv.FormatFloat(l.Load15, 'f', 2, 64),
	}
	j, _ := json.Marshal(loadAvg)

	return j
}

func Mem() []byte {
	v, _ := mem.VirtualMemory()

	// almost every return value is a struct
	//fmt.Printf("Total: %v, Free:%v, UsedPercent:%f%%\n", v.Total, v.Free, v.UsedPercent)

	// convert to JSON. String() is also implemented
	//fmt.Println(v)

	var mem map[string]interface{}
	total := bytesize.New(float64(v.Total))
	available := bytesize.New(float64(v.Available))
	used := bytesize.New(float64(v.Used))
	bytesize.Format = "%.1f "
	usedPercent := math.Round(v.UsedPercent*10) / 10
	//fmt.Printf("%v %v %v", total, free, used)
	mem = map[string]interface{}{
		"total":       total.String(),
		"available":   available.String(),
		"used":        used.String(),
		"usedPercent": uint(usedPercent),
	}
	j, _ := json.Marshal(mem)
	return j
}

func Disk() []byte {
	p, _ := disk.Partitions(true)

	// convert to JSON. String() is also implemented
	//fmt.Println(p)

	var disks []map[string]interface{}
	for _, v := range p {
		//b, _ := json.Marshal(v)
		d, _ := disk.Usage(v.Mountpoint)
		total := bytesize.New(float64(d.Total))
		free := bytesize.New(float64(d.Free))
		used := bytesize.New(float64(d.Used))
		bytesize.Format = "%.1f "
		usedPercent := math.Round(d.UsedPercent*10) / 10
		//fmt.Printf("%v %v %v", total, free, used)
		di := map[string]interface{}{
			"name":        d.Path,
			"total":       total.String(),
			"free":        free.String(),
			"used":        used.String(),
			"usedPercent": uint(usedPercent),
		}
		disks = append(disks, di)
	}

	j, _ := json.Marshal(disks)
	//fmt.Println(string(j))
	return j
}

func Process(pid int32) []byte {
	p, _ := process.NewProcess(pid)

	var proc map[string]interface{}
	name, _ := p.Name()
	cpupercent, _ := p.CPUPercent()
	cpupercent = cpupercent * 100
	cputime, _ := p.Times()
	memory, _ := p.MemoryInfo()
	cmdline, _ := p.Cmdline()
	createtime, _ := p.CreateTime()
	isexists, _ := process.PidExists(pid)
	statuses, _ := p.Status()
	status := strings.Join(statuses, ", ")
	parent, _ := p.Parent()
	ppid, _ := p.Ppid()
	children, _ := p.Children()
	var cnames []string
	var cpids []int32
	var ccmdline []string
	for _, c := range children {
		cn, _ := c.Name()
		ccmd, _ := c.Cmdline()
		cnames = append(cnames, cn)
		cpids = append(cpids, c.Pid)
		ccmdline = append(ccmdline, ccmd)
	}
	fmt.Printf("%v %v %v\n", cnames, cpids, ccmdline)

	//fmt.Printf("%v %v %v %v", name, memory, isexists, status)

	proc = map[string]interface{}{
		"name":       name,
		"cpuPercent": math.Round(cpupercent*10) / 10,
		"cpuTotal":   math.Round(cputime.Total()*100) / 100,
		"cpuUser":    cputime.User,
		"cpuSystem":  cputime.System,
		"cpuIdle":    cputime.Idle,
		"cpuIowait":  cputime.Iowait,
		"vms":        bytesize.New(float64(memory.VMS)).String(),
		"rss":        bytesize.New(float64(memory.RSS)).String(),
		"swap":       bytesize.New(float64(memory.Swap)).String(),
		"cmdline":    cmdline,
		"createTime": time.Unix(createtime/1000, 0).Format(timeformat),
		"isExists":   isexists,
		"status":     status,
		"pid":        p.Pid,
		"parent":     parent,
		"ppid":       ppid,
	}
	j, _ := json.Marshal(proc)
	return j
}
