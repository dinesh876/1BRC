package main

import (
	"github.com/edsrzf/mmap-go"
    "fmt"
    "os"
    "sort"
    "io"
    "runtime"
)

const (
	filePath = "data/measurements.txt"
)

type MemChunk struct{
    start int
    end int
}

type Stats struct {
    min int
    max int
    sum int64
    count int
}

func splitMem(mem mmap.MMap, nSplit int) []MemChunk{
    total := len(mem)
    chunkSize := total / nSplit
    chunks := make([]MemChunk,nSplit)

    chunks[0].start = 0
    for i := 1;i < nSplit; i++ {
        for j := i*chunkSize; j < i*chunkSize + 100; j ++ {
            if mem[j] == '\n'{
                chunks[i-1].end = j
                chunks[i].start = j+1
                break
            }
        }
    }
    chunks[nSplit-1].end = total -1 
    return chunks
}


func readChunks(data mmap.MMap,start int,end int,resultCh chan<- map[string]*Stats){
    station := ""
    temperature := 0
    prev := start
	statsMap := make(map[string]*Stats)   

    for i := start; i < end; i++{
        if data[i] == ';'{
            station = string(data[prev:i])
            i += 1
            temperature = 0
            negative := false
            for data[i] != '\n'{
                ch := data[i]
                if ch == '.'{
                    i+=1
                    continue
                }
                if ch == '-'{
                    negative = true
                    i += 1
                    continue
                }
                ch -= '0'
                if ch > 9 {
                    panic("invalid character")
                }

                temperature = (temperature * 10) + int(ch)
                i+=1
            }
            if negative {
                temperature = -temperature
            }
            
			stats , ok := statsMap[station]
        	if !ok {
            	stats = &Stats{
                	min: temperature,
                	max: temperature,
                	sum: int64(temperature),
                	count: 1,
            	}
        	} else {
            	if temperature < stats.min {
                	stats.min = temperature
            	}
            	if temperature > stats.max {
                	stats.max = temperature
            	}
            	stats.sum += int64(temperature)
            	stats.count++
        	}
        	statsMap[station] = stats
			station = ""
            prev = i + 1
        }
    }
    resultCh <- statsMap
}

func main() {
    maxGoroutines := min(runtime.NumCPU(),runtime.GOMAXPROCS(0))
    fmt.Println("MAXGOROUTINES: ",maxGoroutines)
    f, _ := os.OpenFile(filePath, os.O_RDWR, 0644)
	defer f.Close()
	
    data, err := mmap.Map(f, mmap.RDONLY, 0 )
    if err != nil {
        panic(err)
    }
	defer data.Unmap()

    chunks := splitMem(data,maxGoroutines)
    resultCh := make(chan map[string]*Stats)
    totals := make(map[string]*Stats)
    for i := 0;i < maxGoroutines; i ++ {
        //fmt.Printf("Chunk %d: [%d:%d]\n",idx,chunk.start,chunk.end)
        go readChunks(data,chunks[i].start,chunks[i].end,resultCh)
    }

    for i := 0;i < maxGoroutines; i ++ {
        results := <- resultCh
        for station,stats := range results{
            total := totals[station]
            if total == nil {
                totals[station] = stats
            } else {
                total.min  = min(total.min,stats.min)
                total.max = min(total.max,stats.max)
                total.sum += stats.sum
                total.count += stats.count
            }
        }
    }

	printResult(os.Stdout,totals)
}


func printResult(out io.Writer,data map[string]*Stats){
	stations := make([]string,0,len(data))
    
    for station := range data{
        stations = append(stations,station)
    }

    sort.Strings(stations)
    fmt.Fprintf(out,"{")
    for idx,key := range stations{
        if idx > 0 {
            fmt.Fprintf(out,", ")
        }
        s := data[key]
        mean := float64(s.sum/10) / float64(s.count)
        min := float64(s.min) / 10
        max := float64(s.max) / 10
        fmt.Fprintf(out,"%s=%.1f/%.1f/%.1f",key,min,max,mean)
    }
    fmt.Fprintf(out,"}\n")
}
