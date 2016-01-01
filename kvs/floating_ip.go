package kvs

// FipKVS is a floating ip management kvs.
type FipKVS struct {
	KVS
}

// NewFipKVS builds a FipKVS instance.
func NewFipKVS(backend KVS) *FipKVS {
	return &FipKVS{
		KVS: backend,
	}
}
