package ymdsnowflake

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/deckarep/golang-set"
)

var sf *YMDSnowflake

var startTime int64
var srvID int64
var ymdpre int64

func init() {

	srvID = 1
	sf = NewYMDSnowflake(1)
	if sf == nil {
		panic("Snowflake not created")
	}

	startTime = 0
	ymdpre = sf.ymd*DataUnit

}

func nextID(t *testing.T) int64 {
	id := sf.NextID()

	return id
}

func Decompose(id int64) map[string]int64 {
	const maskSequence = int64((1<<BitLenSequence - 1) << BitLenSrvID)
	const maskMachineID = int64(1<<BitLenSrvID - 1)

	msb := id >> 63
	time := (id-(ymdpre)) >> (BitLenSequence + BitLenSrvID)
	sequence := (id-(ymdpre)) & maskSequence >> BitLenSrvID
	srvID := id & maskMachineID
	return map[string]int64{
		"id":         id,
		"msb":        msb,
		"time":       time,
		"sequence":   sequence,
		"srv-id": srvID,
	}
}



func TestSnowflakeOnce(t *testing.T) {
	//sleepTime := uint64(50)
	//time.Sleep(time.Duration(sleepTime) * 10 * time.Millisecond)

	id := nextID(t)
	parts := Decompose(id)

	actualMSB := parts["msb"]
	if actualMSB != 0 {
		t.Errorf("unexpected msb: %d", actualMSB)
	}

	actualTime := parts["time"]
	if actualTime < 0 {
		t.Errorf("unexpected time: %d", actualTime)
	}

	//if actualTime < sleepTime || actualTime > sleepTime+1 {
	//	t.Errorf("unexpected time: %d", actualTime)
	//}

	actualSequence := parts["sequence"]
	if actualSequence != 0 {
		t.Errorf("unexpected sequence: %d", actualSequence)
	}

	actualMachineID := parts["srv-id"]
	if actualMachineID != srvID {
		t.Errorf("unexpected machine id: %d", actualMachineID)
	}

	fmt.Println("Snowflake id:", id)
	fmt.Println("decompose:", parts)
}

func currentTime() int64 {
	now := time.Now()
	y,m,d := now.Date()

	zero := time.Date(y,m,d,0,0,0,0,now.Location())
	return (now.Sub(zero).Nanoseconds()/snowflakeTimeUnit)
}

func TestSnowflakeFor10Sec(t *testing.T) {
	var numID uint32
	var lastID int64
	var maxSequence int64

	initial := currentTime()
	current := initial
	for current-initial < 1000 {
		id := nextID(t)
		parts := Decompose(id)
		numID++

		if id <= lastID {
			lastPart := Decompose(lastID)
			t.Fatalf("duplicated id, ID :%d, lastID:%d, part:%v, lastpart:%v \n",id,lastID,parts,lastPart)
		}
		lastID = id

		current = currentTime()

		actualMSB := parts["msb"]
		if actualMSB != 0 {
			t.Errorf("unexpected msb: %d", actualMSB)
		}

		actualTime := int64(parts["time"])
		overtime := startTime + actualTime - current
		if overtime > 0 {
			t.Errorf("unexpected overtime: %d, actualTime： %d", overtime, actualTime)
		}

		actualSequence := parts["sequence"]
		if maxSequence < actualSequence {
			maxSequence = actualSequence
		}

		actualMachineID := parts["srv-id"]
		if actualMachineID != srvID {
			t.Errorf("unexpected machine id: %d", actualMachineID)
		}
	}

	if maxSequence != 1<<BitLenSequence-1 {
		t.Logf("unexpected max sequence: %d", maxSequence)
	}
	fmt.Println("max sequence:", maxSequence)
	fmt.Println("number of id:", numID)
}

func TestSnowflakeInParallel(t *testing.T) {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	fmt.Println("number of cpu:", numCPU)

	consumer := make(chan int64)

	const numID = 10000000
	generate := func() {
		for i := 0; i < numID; i++ {
			consumer <- nextID(t)
		}
	}

	const numGenerator = 10
	for i := 0; i < numGenerator; i++ {
		go generate()
	}

	set := mapset.NewSet()
	for i := 0; i < numID*numGenerator; i++ {
		id := <-consumer
		if set.Contains(id) {
			t.Log("number of id:", set.Cardinality())
			t.Fatal("duplicated id")

		} else {
			set.Add(id)
		}
	}
	fmt.Println("number of id:", set.Cardinality())
}



