package queops

import "github.com/Workiva/go-datastructures/queue"

type pqueueItem struct {
	priority  int8
	StructPtr *interface{}
}

func (pqi pqueueItem) Compare(other queue.Item) int {
	opqi := other.(pqueueItem)
	if pqi.priority > opqi.priority {
		return 1
	} else if pqi.priority == opqi.priority {
		return 0
	}
	return -1
}

//Comment - "github.com/Workiva/go-datastructures/queue"

//var WGroup sync.WaitGroup

func InitializePriorityQueue() *queue.PriorityQueue {
	return queue.NewPriorityQueue(1, false)
}
func AddToPriorityQue(PQueue *queue.PriorityQueue, item *interface{}, priority int8) error {

	//pqi := pqueueItem{priority, item}
	return PQueue.Put(pqueueItem{priority, item})

}
func ReadFromPriorityQue(PQueue *queue.PriorityQueue) (pqueueItem, error) {

	item, err := PQueue.Get(1)

	item3 := item[0].(pqueueItem)
	return item3, err

}

/*func BenchmarkPriorityQueue(b *testing.B) {
	q := NewPriorityQueue(b.N, false)

	wg.Add(1)
	i := 0

	go func() {
		for {
			q.Get(1)
			i++
			if i == b.N {
				wg.Done()
				break
			}
		}
	}()

	for i := 0; i < b.N; i++ {
		q.Put(mockItem(i))
	}

	wg.Wait()
}
*/
