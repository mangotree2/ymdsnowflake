package ymdsnowflake

import (
	"sync"
	"time"
)



type YMDSnowflake struct {
	mutex       *sync.Mutex
	sequence    uint16
	srvID   	uint16
	elapsedTime	int64
	ymd			int64
}

const snowflakeTimeUnit = int64(1e6) // nsec, i.e. 1 msec
const DataUnit = 1e13 // 190228
const BitLenSequence  = 12  // bit length of sequence number
const BitLenSrvID  = 4  // bit length of srvID


func (sf *YMDSnowflake)NextID() int64 {

	const maskSequence = uint16(1<<BitLenSequence - 1)

	sf.mutex.Lock()
	defer sf.mutex.Unlock()

	now := time.Now()
	y,m,d := now.Date()

	zero := time.Date(y,m,d,0,0,0,0,now.Location())

	//zy,zm,zd := zero.Date()
	//fmt.Println("y,m,d:", y,m,d,"zero :",zy,zm,zd)
	//fmt.Println("na : ",now.Sub(zero).Nanoseconds(),"ms : ",now.Sub(zero).Nanoseconds()/snowflakeTimeUnit)
	//fmt.Printf("%d\n",ymd*DataUnit)

	var ymd = int64(y%100*1e4+int(m)*1e2+d)

	if sf.ymd < ymd {
		sf.sequence = uint16(1<<BitLenSequence - 1)
		sf.elapsedTime = 0
		sf.ymd = ymd
	}

	current := now.Sub(zero).Nanoseconds() / snowflakeTimeUnit
	if sf.elapsedTime < current {
		sf.elapsedTime = current
		sf.sequence = 0
	} else {
		sf.sequence = (sf.sequence + 1) & maskSequence
		if sf.sequence == 0 {
			sf.elapsedTime++
			overtime := sf.elapsedTime - current
			time.Sleep(sleepTime((overtime)))
		}
	}


	return ymd*DataUnit + ((now.Sub(zero).Nanoseconds()/snowflakeTimeUnit) <<(BitLenSrvID+BitLenSequence)|
		int64(sf.sequence) << BitLenSrvID | int64(sf.srvID))

}

//非线程安全
func (sf *YMDSnowflake) getYMD() int64 {
	return sf.ymd
}

func getYMD() int64 {
	now := time.Now()
	y,m,d := now.Date()
	return int64(y%100*1e4+int(m)*1e2+d)
}

func getYMDZero(now time.Time) time.Time {
	y,m,d := now.Date()
	return time.Date(y,m,d,0,0,0,0,now.Location())
}

func sleepTime(overtime int64) time.Duration {
	return time.Duration(overtime)*10*time.Millisecond -
		time.Duration(time.Now().UTC().UnixNano()%snowflakeTimeUnit)*time.Nanosecond
}

func NewYMDSnowflake(srvID uint16) *YMDSnowflake {

	return &YMDSnowflake{
		mutex:       new(sync.Mutex),
		sequence:    uint16(1<<BitLenSequence - 1),
		srvID:       srvID,
		elapsedTime: 0,
		ymd:         getYMD(),
	}
}



