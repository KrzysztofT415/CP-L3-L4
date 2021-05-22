package app

import (
	"math"
	"math/rand"
	"time"
)

type Graph struct {
	ReadersCount  int
	ReaderLeaving chan bool
	ReadersQueue  *ReaderQueue
	WritersQueue  *WriterQueue

	VerticesInfo []*Vertex
}

func NewGraph(n int, d int) *Graph {
	newGraph := new(Graph)
	newGraph.ReadersCount = 0
	newGraph.ReaderLeaving = make(chan bool)
	newGraph.ReadersQueue = MakeUnboundedQueueOfReaders()
	newGraph.WritersQueue = MakeUnboundedQueueOfWriters()

	//CREATING VERTICES WITH BASIC EDGES (i, i + 1)

	newGraph.VerticesInfo = make([]*Vertex, n)
	newGraph.VerticesInfo[0] = NewVertex(0, n, newGraph.ReadersQueue, newGraph.WritersQueue)
	for i := 0; i < n-1; i++ {
		newGraph.VerticesInfo[i+1] = NewVertex(i+1, n, newGraph.ReadersQueue, newGraph.WritersQueue)
		newGraph.VerticesInfo[i].NextVerticesInfo = append(newGraph.VerticesInfo[i].NextVerticesInfo, newGraph.VerticesInfo[i+1])
		newGraph.VerticesInfo[i+1].NextVerticesInfo = append(newGraph.VerticesInfo[i+1].NextVerticesInfo, newGraph.VerticesInfo[i])
	}

	//ADDING RANDOM EDGES - SHORTCUTS (j, k)

	var added [][]int
	for i := 0; i < d; i++ {
		w := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(n - 2)
		v := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(n-w-2) + 2

		var wasUsed = false
		for i := range added {
			if added[i][0] == w && added[i][1] == v {
				wasUsed = true
			}
		}
		if !wasUsed {
			added = append(added, []int{w, v})
			newGraph.VerticesInfo[w].NextVerticesInfo = append(newGraph.VerticesInfo[w].NextVerticesInfo, newGraph.VerticesInfo[w+v])
			newGraph.VerticesInfo[w+v].NextVerticesInfo = append(newGraph.VerticesInfo[w+v].NextVerticesInfo, newGraph.VerticesInfo[w])
			i++
		}
		i--
	}

	//UPDATING ROUTE TABLE

	for _, v := range newGraph.VerticesInfo {
		for j := 0; j < n; j++ {
			cost := int(math.Abs(float64(v.Id - j)))
			next := v.Id + 1
			if j == v.Id {
				cost = -1
				next = -1
			} else if j < v.Id {
				next -= 2
			}
			v.ThisRoutingTable.RoutingInfos[j] = NewRoutingInfo(j, next, cost)
		}
		for _, n := range v.NextVerticesInfo {
			v.ThisRoutingTable.RoutingInfos[n.Id].SourceId = n.Id
			v.ThisRoutingTable.RoutingInfos[n.Id].NextHop = n.Id
			v.ThisRoutingTable.RoutingInfos[n.Id].Cost = 1
		}
	}

	return newGraph
}
