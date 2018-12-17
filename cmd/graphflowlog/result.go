package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/santrancisco/cque"
	"github.com/santrancisco/graphflowlog/jobs"
)

// TODO: Find home for result handler. :)
type Result interface {
	String() string
}

type ResultHandler struct {
	// Interval is the amount of time that this Worker should sleep before trying
	// to find another Job.
	WaitForResult bool
	WaitTimeStart time.Time
	Data          Data
	TempEdges     []string
	qc            *cque.Client
	ctx           context.Context
}

func (rh *ResultHandler) Run() {
	// Process flowlog from io reader stream
	rh.Data = Data{}
	rh.Data.Extent = Extent{
		Xmin: -2172.1,
		Xmax: 1796.1,
		Ymin: -596.1,
		Ymax: 247.1,
		Dx:   3968.1,
		Dy:   844.1,
	}
	rh.Data.Coldistance = COLLUMNDISTANCE
	rh.Data.Nodes = map[string]*Node{}
	rh.TempEdges = []string{}

	for {
		rh.WaitForResult = true
		rh.WaitTimeStart = time.Now()
		select {
		case <-rh.ctx.Done():
			log.Printf("[DEBUG] Result Handler is done\n")
			return
		case r := <-rh.qc.Result:
			result := r.Result.(jobs.Connection)
			rh.Data.AddNode(result.From)
			rh.Data.AddNode(result.To)
			rh.Data.Nodes[result.To].Addport(result.Port)
			if rh.Data.Nodes[result.From].Addtarget(result.To) {
				rh.Data.Nodes[result.To].Addtarget(result.From)
				rh.TempEdges = append(rh.TempEdges, fmt.Sprintf("%s %s", result.From, result.To))
			}

		}
	}
}

// Sample flowlog:
// Ideas:  We only care about records that has FROMPORT > TOPORT
// PUSH FROMIP + TOIP + TOPORT
// TOIP + TOPORT

// 2018-09-24T00:00:00.000Z 2 033604130140 eni-eb841bef 10.65.8.57 10.65.25.164 80 55743 6 5 430 1537747200 1537747201 ACCEPT OK
// 2018-09-24T00:00:00.000Z 2 033604130140 eni-eb841bef 10.65.8.57 10.65.25.164 80 55001 6 5 425 1537747200 1537747201 ACCEPT OK
// 2018-09-24T00:00:00.000Z 2 033604130140 eni-eb841bef 10.65.25.164 10.32.1.18 8132 58060 6 36 48534 1537747200 1537747201 ACCEPT OK
// 2018-09-24T00:00:00.000Z 2 033604130140 eni-eb841bef 10.65.25.164 10.32.1.18 8132 55786 6 3 440 1537747200 1537747201 ACCEPT OK
// 2018-09-24T00:00:00.000Z 2 033604130140 eni-eb841bef 10.65.25.164 10.65.8.57 55161 80 6 5 502 1537747200 1537747201 ACCEPT OK
// 2018-09-24T00:00:00.000Z 2 033604130140 eni-eb841bef 10.65.25.164 10.65.8.57 55350 80 6 5 495 1537747200 1537747201 ACCEPT OK
// 2018-09-24T00:00:00.000Z 2 033604130140 eni-eb841bef 10.65.8.57 10.65.25.164 80 55654 6 36 46056 1537747200 1537747201 ACCEPT OK
// 2018-09-24T00:00:00.000Z 2 033604130140 eni-eb841bef 10.65.25.164 52.200.21.20 55365 443 6 18 13971 1537747200 1537747201 ACCEPT OK

const (
	// Distance between ip range collumn
	COLLUMNDISTANCE = 50
	ROWDISTANCE     = 2
	TARGETMAXWEIGHT = 30.0
)

var xaxisloookup []int64

type Graphics struct {
	D    int64  `json:"d"`
	H    int64  `json:"h"`
	W    int64  `json:"w"`
	X    int64  `json:"x"`
	Y    int64  `json:"y"`
	Z    int64  `json:"z"`
	Fill string `json:"fill"`
}
type Node struct {
	Color     int64    `json:"color"`
	Graphics  Graphics `json:"graphics"`
	Noisyrank int64    `json:"noisyrank"`
	Label     string   `json:"label"`
	Targets   []string `json:"targets"`
	Ports     []int64  `json:ports`
}

type Extent struct {
	Xmin float32 `json:"xmin"`
	Xmax float32 `json:"xmax"`
	Ymin float32 `json:"ymin"`
	Ymax float32 `json:"ymax"`
	Dx   float32 `json:"dx"`
	Dy   float32 `json:"dy"`
}

/*
Example of data.json for our app:
{
	"extent": {
	"xmin": -2172.5603,
	"xmax": 1796.19,
	"ymin": -596.4198,
	"ymax": 247.91452,
	"dx": 3968.7503,
	"dy": 844.33432
	},
	"nodes": {
		"10.10.1.2": {
			"graphics": {
				"d": 10,
				"h": 10,
				"w": 10,
				"x": 519.0851,
				"y": 115.714005,
				"z": 0,
			},
			"noisyrank": "0.224842371677",
			"label": "10.10.1.2",
			"targets": [
				"10.10.10.1",
				"10.10.1.1"
			]
		}
	},
	,"nodeIndex": [
        [
            519.0851,
            115.714005,
            519.0851,
            115.714005,
            0.9
		]
	],
	"edgeIndex": [
        [
            519.0851,
            115.714005,
            0,
            0,
            0.85692084
		]
	]
}
*/

type Data struct {
	Extent       Extent           `json:"extent"`
	Nodes        map[string]*Node `json:"nodes"`
	NodeIndex    [][]int64        `json:"nodeIndex"`
	EdgeIndex    [][]int64        `json:"edgeIndex"`
	Xaxisloookup []int64          `json:"xaxisloookup"`
	Coldistance  int              `json:"coldistance"`
}

func (data *Data) AddNodeIndex() {
	for _, n := range data.Nodes {
		if len(n.Targets) > 30 {
			n.Noisyrank = 100
		} else {
			n.Noisyrank = int64(len(n.Targets)) * 100 / 30
		}
		data.NodeIndex = append(data.NodeIndex, []int64{
			n.Graphics.X,
			n.Graphics.Y,
			n.Graphics.X,
			n.Graphics.Y,
			n.Noisyrank,
		})
	}
}

func (data *Data) Addtoxaxislookup(ip string) {
	i := iptoint(ip)
	xlookup, _ := getaxis(i)
	for _, v := range data.Xaxisloookup {
		if xlookup == v {
			return
		}
	}
	data.Xaxisloookup = append(data.Xaxisloookup, xlookup)
}

func (data *Data) AddEdgeIndexes(connections []string) {
	for _, c := range connections {
		ips := strings.Split(c, " ")
		data.EdgeIndex = append(data.EdgeIndex, []int64{
			data.Nodes[ips[0]].Graphics.X,
			data.Nodes[ips[0]].Graphics.Y,
			data.Nodes[ips[1]].Graphics.X,
			data.Nodes[ips[1]].Graphics.Y,
			50, // This is the weight for our edge (0 to 100)
		})
	}
}

func (data *Data) Findnodefromip(ip string) *Node {
	for _, node := range data.Nodes {
		if node.Label == ip {
			return node
		}
	}
	return nil
}

func (data *Data) Updatexasis() {
	sort.Slice(data.Xaxisloookup, func(i, j int) bool {
		return data.Xaxisloookup[i] < data.Xaxisloookup[j]
	})
	for ip, node := range data.Nodes {
		i := iptoint(ip)
		l := len(data.Xaxisloookup)
		xlookup, y := getaxis(i)
		for i, v := range data.Xaxisloookup {
			if xlookup == v {
				node.Graphics.X = int64((l/2 - i) * COLLUMNDISTANCE)
				break
			}
		}
		node.Graphics.Y = y * ROWDISTANCE
	}
}

func Newnode(ip string) *Node {
	g := Graphics{
		D:    10,
		H:    10,
		W:    10,
		X:    0,
		Y:    0,
		Z:    0,
		Fill: "#999999",
	}

	return &Node{
		Graphics:  g,
		Label:     ip,
		Ports:     []int64{},
		Targets:   []string{},
		Noisyrank: 0.0,
		Color:     16777212,
	}
}

func (node *Node) Addtarget(ip string) bool {
	for _, i := range node.Targets {
		if i == ip {
			return false
		}
	}
	node.Targets = append(node.Targets, ip)
	return true

}

func (node *Node) Addport(port int64) {
	for _, i := range node.Ports {
		if i == port {
			return
		}
	}
	node.Color = (node.Color + port*300) % 16777215
	node.Ports = append(node.Ports, port)
}

// ip - mod give us the integer value of the base ip for /24 range of this ip - we use this to lookup the actual xaxis
// (mod - 128)* ROWDISTANCE : Assuming 0 position on y-axis is at 128, we use this to calculate the Y axis location for this node.
func getaxis(ip int64) (x, y int64) {
	mod := ip % 256
	return (ip - mod), (mod - 128) * ROWDISTANCE
}

// Convert ip string to an integer
func iptoint(ips string) int64 {
	ip := net.ParseIP(ips)
	IPv4Int := big.NewInt(0)
	IPv4Int.SetBytes(ip.To4())
	return IPv4Int.Int64()
}

// Check If ip is not between internal ranges
// 	10.0.0.0 – 10.255.255.255
//  172.16.0.0 – 172.31.255.255
//  192.168.0.0 – 192.168.255.255
func Isinternal(ip string) bool {
	i := iptoint(ip)
	if !(((i > 167772160) && (i < 184549375)) || ((i > 2886729728) && (i < 2887778303)) || ((i > 3232235520) && (i < 3232301055))) {
		return false
	}
	return true
}

func (data *Data) AddNode(ip string) bool {
	// Bail if we find existing node
	node := data.Findnodefromip(ip)
	if node != nil {
		return true
	}
	data.Nodes[ip] = Newnode(ip)
	data.Addtoxaxislookup(ip)
	return true
}
