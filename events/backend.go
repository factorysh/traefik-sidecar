package events

import "reflect"

type Server struct {
	Url    string `json:",omitempty"`
	Weight int    `json:",omitempty"`
}

type Backend struct {
	LoadBalancer interface{}        `json:"loadBalancer"`
	Servers      map[string]*Server `json:"servers"`
}

type Backends map[string]*Backend

// Diff before after
func Diff(b1, b2 Backends) Backends {
	r := make(Backends)
	for b, v1 := range b1 {
		v2, ok := b2[b]
		if ok {
			if !reflect.DeepEqual(v1, v2) {
				r[b] = v2
			}
		} else {
			r[b] = v2
		}
	}
	for b := range b2 {
		_, ok := b1[b]
		if !ok {
			r[b] = nil
		}
	}
	return r
}
