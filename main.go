package main

import (
    "fmt"
    "os"
    "io"
    "bufio"
    "strings"
    "strconv"
    "sort"
    "runtime"
)

const (
    file_path = "data/measurements.txt"
)

type Stats struct {
    min float64 
    max float64
    sum float64
    count int
}

func main(){
    fmt.Printf("Program is configured to use %d CPU cores.\n", runtime.GOMAXPROCS(-1))
    statsMap := make(map[string]*Stats)
    f,err := os.Open(file_path)
    if err != nil {
        panic(err)
    }
    
    defer f.Close()

    scanner := bufio.NewScanner(f)
    for scanner.Scan(){
        line := scanner.Text()
        stationName,temperature,_:=  strings.Cut(line,";")
        
        temp,err := strconv.ParseFloat(temperature,64)
        if err != nil {
            fmt.Println("Failed to parse the temperature:",temperature)
            continue
        }



        stats , ok := statsMap[stationName] 
        if !ok {
            stats = &Stats{
                min: temp,
                max: temp,
                sum: temp,
                count: 1,
            }
        } else {
            if temp < stats.min {
                stats.min = temp
            }
            if temp > stats.max {
                stats.max = temp
            }
            stats.sum += temp
            stats.count++
        }
        statsMap[stationName] = stats
    }

    if err := scanner.Err();err != nil{
        fmt.Println("Error reading file: %s",err)
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
        mean := s.sum / float64(s.count)
        fmt.Fprintf(out,"%s=%.1f/%.1f/%.1f",key,s.min,s.max,mean)
    }
    fmt.Fprintf(out,"}\n")
}
