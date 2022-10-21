package main

type HistoryNode struct {
	Record   byte
	Previous *HistoryNode
}

type HistoryStack struct {
	Top    *HistoryNode
	Length int
}

type History struct {
	Fg *HistoryStack
	Bg *HistoryStack
}

func (h *HistoryStack) add(b byte) {
	newTop := &HistoryNode{
		Record:   b,
		Previous: h.Top,
	}
	h.Top = newTop
	h.Length++
}

func (h *HistoryStack) pop() byte {
	if h == nil || h.Top == nil {
		return 0
	}
	h.Top = h.Top.Previous
	h.Length--
	return h.Top.Record
}

func (h *HistoryStack) top() byte {
	if h == nil || h.Top == nil {
		return 0
	}
	return h.Top.Record

}

func (h *HistoryStack) back(n int) byte {
	if h == nil || h.Top == nil {
		return 0
	}
	node := h.Top
	for n > 0 {
		node = node.Previous
		n--
	}
	if node == nil {
		return 0
	}
	return node.Record
}

func (h *HistoryStack) toArr() []byte {
	if h == nil {
		return nil
	}
	if h.Length == 0 {
		h.Length = 0
	}
	stack := make([]byte, h.Length)
	node := h.Top
	for i := h.Length - 1; i >= 0; i-- {
		if node.Record == 0 {
			return stack
		}
		stack[i] = node.Record
		node = node.Previous
	}
	return stack
}

func toHistory(records []byte) *HistoryStack {
	stack := &HistoryStack{Top: nil}
	for _, record := range records {
		stack.add(record)
	}
	return stack
}

func (h *HistoryStack) removeLastMatching(match byte) {
	if h == nil || h.Top == nil || h.Top.Record == 0 {
		return
	}
	node := h.Top
	if node.Record == match {
		h.Top = node.Previous
		return
	}

	for node.Previous != nil {
		if node.Previous.Record == match {
			node.Previous = node.Previous.Previous
			return
		}
	}
}

func (h *History) removeFirstMatch(val byte, fg bool) {
	if fg {
		h.Fg.removeLastMatching(val)
	} else {
		h.Bg.removeLastMatching(val)
	}
}
