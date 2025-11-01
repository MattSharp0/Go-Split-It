package heaps

import (
	"github.com/MattSharp0/transaction-split-go/internal/models"
)

type MaxNetBalanceHeap []*models.NetBalance

func (h MaxNetBalanceHeap) Len() int           { return len(h) }
func (h MaxNetBalanceHeap) Less(i, j int) bool { return h[i].NetBalance.GreaterThan(h[j].NetBalance) }
func (h MaxNetBalanceHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *MaxNetBalanceHeap) Push(x any) {
	*h = append(*h, x.(*models.NetBalance))
}

func (h *MaxNetBalanceHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type MinNetBalanceHeap []*models.NetBalance

func (h MinNetBalanceHeap) Len() int           { return len(h) }
func (h MinNetBalanceHeap) Less(i, j int) bool { return h[i].NetBalance.LessThan(h[j].NetBalance) }
func (h MinNetBalanceHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *MinNetBalanceHeap) Push(x any) {
	*h = append(*h, x.(*models.NetBalance))
}

func (h *MinNetBalanceHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
