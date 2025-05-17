package snowflake

import (
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (

	// Epoch is set to the twitter snowflake epoch of Nov 04 2010 01:42:54 UTC in milliseconds
	// You may customize this to set a different epoch for your application.
	epoch int64 = 1288834974657

	// NodeBits holds the number of bits to use for Node
	// Remember, you have a total 22 bits to share between Node/Step
	nodeBits uint8 = 10

	// StepBits holds the number of bits to use for Step
	// Remember, you have a total 22 bits to share between Node/Step
	stepBits uint8 = 12

	nodes map[int64]*node
	mu    sync.Mutex

	nodeMax   = -1 ^ (-1 << nodeBits)
	nodeMask  = nodeMax << stepBits
	stepMask  = -1 ^ (-1 << stepBits)
	timeShift = nodeBits + stepBits
	nodeShift = stepBits

	base62Chars = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
)

type ID int64

type node struct {
	mu     sync.Mutex
	epoch  time.Time
	time   int64
	nodeNo int64
	step   int64

	nodeMax   int64
	nodeMask  int64
	stepMask  int64
	timeShift uint8
	nodeShift uint8
}

func init() {
	nodes = make(map[int64]*node)

}

func Generate(nodeNo int64) (ID, error) {
	node, err := getNode(nodeNo)
	if err != nil {
		return 0, err
	}
	id := node.Generate()

	return id, nil
}

func getNode(nodeNo int64) (*node, error) {
	if nodes[nodeNo] != nil {
		return nodes[nodeNo], nil
	}

	mu.Lock()
	defer mu.Unlock()
	if nodes[nodeNo] != nil {
		return nodes[nodeNo], nil
	}
	var err error
	nodes[nodeNo], err = newNode(nodeNo)

	return nodes[nodeNo], err
}

// NewNode returns a new snowflake node that can be used to generate snowflake
// IDs
func newNode(nodeNo int64) (*node, error) {

	//if nodeBits+stepBits > 22 {
	//	return nil, errors.New("Remember, you have a total 22 bits to share between Node/Step")
	//}

	n := node{}
	n.nodeNo = nodeNo
	n.nodeMax = -1 ^ (-1 << nodeBits)
	n.nodeMask = n.nodeMax << stepBits
	n.stepMask = -1 ^ (-1 << stepBits)
	n.timeShift = nodeBits + stepBits
	n.nodeShift = stepBits

	if n.nodeNo < 0 || n.nodeNo > n.nodeMax {
		return nil, errors.New("Node number must be between 0 and " + strconv.FormatInt(n.nodeMax, 10))
	}

	var curTime = time.Now()
	// add time.Duration to curTime to make sure we use the monotonic clock if available
	n.epoch = curTime.Add(time.Unix(epoch/1000, (epoch%1000)*1000000).Sub(curTime))

	return &n, nil
}

// Generate creates and returns a unique snowflake ID
// To help guarantee uniqueness
// - Make sure your system is keeping accurate system time
// - Make sure you never have multiple nodes running with the same node ID
func (n *node) Generate() ID {

	now := time.Now().UTC().Sub(n.epoch).Milliseconds()
	currentNodeTime := n.time

	for i := 0; i < 10; i++ {

		if now == currentNodeTime {
			step := atomic.AddInt64(&n.step, 1) & n.stepMask

			if step == 0 {
				time.Sleep(time.Millisecond)
				continue
			}
			break
		} else {
			n.mu.Lock()
			defer n.mu.Unlock()
			if now == n.time {
				continue
			}
			n.step = 0
			currentNodeTime = now
			n.time = now
			break
		}
	}

	r := ID((now)<<n.timeShift |
		(n.nodeNo << n.nodeShift) |
		(n.step),
	)

	return r
}

func (id ID) EncodeBase62() string {
	if id == 0 {
		return "0"
	}

	result := []byte{}
	for id > 0 {
		result = append([]byte{base62Chars[id%62]}, result...)
		id = id / 62
	}
	return string(result)
}
