package mirror

type Middleware func(next HandleFunc) HandleFunc
