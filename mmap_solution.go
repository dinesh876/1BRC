package main

import (
	"github.com/edsrzf/mmap-go"
    "fmt"
    "os"
    "sort"
    "io"
)

const (
	filePath = "data/measurements.txt"
)


type Stats struct {
    min int
    max int
    sum int64
    count int
}


func main() {
	f, _ := os.OpenFile(filePath, os.O_RDWR, 0644)
	defer f.Close()
	
    data, err := mmap.Map(f, mmap.RDONLY, 0 )
    if err != nil {
        panic(err)
    }
	defer data.Unmap()

    station := ""
    temperature := 0
    total := len(data)
    prev := 0
	statsMap := make(map[string]*Stats)   

    for i := 0; i < total; i++{
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
	printResult(os.Stdout,statsMap)
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
