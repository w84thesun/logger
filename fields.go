package logger

import "sync"

type Fields map[string]interface{}

func (f Fields) Copy() Fields {
	newF := make(Fields, len(f))
	for k, v := range f {
		newF[k] = v
	}
	return newF
}

func (f Fields) Merge(newValues Fields) Fields {
	copied := f.Copy()

	for k, v := range newValues {
		copied[k] = v
	}
	return copied
}

//nolint
var ignore = map[string]struct{}{
	"@timestamp": {},
	"message":    {},
	"level":      {},
	"service":    {},
}

// Flattens map to loosely coupled k-v pairs to pass into .With
func (f Fields) Flatten() []interface{} {
	list := flattenPool.Get().([]interface{})

	for k, v := range f {
		if _, ok := ignore[k]; ok {
			continue
		}
		list = append(list, k, v)
	}

	return list
}

var flattenPool = sync.Pool{
	New: func() interface{} {
		return []interface{}{}
	},
}

func putFlatten(flatten []interface{}) {
	//nolint
	flattenPool.Put(flatten[:0])
}
