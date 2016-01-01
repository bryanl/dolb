package kvs

import "fmt"

// MkdirError is an a kvs mkdir error.
type MkdirError struct {
	Dir string
	Err error
}

func (mde *MkdirError) Error() string {
	return fmt.Sprintf("could not create %q directory: %v", mde.Dir, mde.Err)
}

// KVError is a general kvs error.
type KVError struct {
	Key string
	Err error
}

func (ke *KVError) Error() string {
	return fmt.Sprintf("could not set %q: %v", ke.Key, ke.Err)
}

// KVDeleteError is a error while deleting an entry from the kvs.
type KVDeleteError struct {
	Key string
	Err error
}

func (kde *KVDeleteError) Error() string {
	return fmt.Sprintf("could not delete %q: %v", kde.Key, kde.Err)
}
