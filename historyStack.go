package main

type HistoryNode struct {
	Record * TilePair
	Previous * HistoryNode
}

type HistoryStack struct {
	Top * HistoryNode
	Length int
}

func (h * HistoryStack) add(t * TilePair) {
	newTop := &HistoryNode{
		Record:   t,
		Previous: h.Top,
	}
	h.Top = newTop
	h.Length++
}

func (h * HistoryStack) pop() * TilePair {
	if h == nil || h.Top == nil {
		return nil
	}
	h.Top = h.Top.Previous
	h.Length--
	return h.Top.Record
}

func (h * HistoryStack) top() * TilePair {
	if h == nil || h.Top == nil{
		return nil
	}
	return h.Top.Record

}


func (h * HistoryStack) back(n int) * TilePair {
	if h == nil || h.Top == nil {
		return nil
	}
	node := h.Top
	for n > 0 {
		node = node.Previous
		n--
	}
	if node == nil {
		return nil
	}
	return node.Record
}

func (h * HistoryStack) toArr() [] * TilePair {
	if h == nil {
		return nil
	}
	if h.Length == 0 {
		h.Length = 0
	}
	stack := make([] * TilePair, h.Length)
	node := h.Top
	for i := h.Length - 1; i >= 0 ; i -- {
		if node.Record == nil {
			return stack
		}
		stack[i] = node.Record
		node = node.Previous
	}
	return stack
}

func toHistory (records [] * TilePair) * HistoryStack {
	stack := &HistoryStack{Top: nil}
	for _,record := range records{
		stack.add(record)
	}
	return stack
}

func (h * HistoryStack) removeFirstMatch(val byte, fg bool) {
	if h == nil || h.Top == nil  || h.Top.Record == nil{
		return
	}
	var prev * HistoryNode
	node := h.Top
	for true {
		if (fg && node.Record.ShownSymbol == val) || (!fg && node.Record.BackgroundCode == val) {
			if prev != nil {
				prev.Previous = node.Previous
			}else {
				h.Top = node.Previous
			}
			return
		}
		if node.Previous == nil || node.Previous.Record == nil {
			return
		}
		prev = node
		node = node.Previous
	}

}
