package dolbutil_test

import (
	"math/rand"
	"sync"

	. "github.com/bryanl/dolb/dolbutil"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Random", func() {
	It("is concurrency safe", func() {
		rnd := rand.New(NewSource())
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				rnd.Int63()
				wg.Done()
			}()
		}
		wg.Wait()
	})
})
