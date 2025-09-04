package main


type stats strct {
    min float64
    max float64
    sum float64
    count int
}

type store struct{
    kv map[string]stats
    mu sync.Mutex
}


func NewStore() *store{
    return &store{
        kv: make(map[string]stats{})
    }
}
