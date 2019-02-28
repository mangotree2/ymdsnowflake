# ymdsnowflake


```go

// eg : today is 20190228; 
// use NextID will return int64(1902284193284128769)
/*

| year-month-day|  time   | sequence |  srvID or MachineID|
|               | 27bit   | 12bit    |  4bit              | 
*/
 
sf = NewYMDSnowflake(1)
id := sf.NextID()



```