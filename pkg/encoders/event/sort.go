package event

// Ascending is a slice of events that sorts in chronological order (oldest
// first.
type Ascending []*E

func (ev Ascending) Len() int           { return len(ev) }
func (ev Ascending) Less(i, j int) bool { return ev[i].CreatedAt.I64() < ev[j].CreatedAt.I64() }
func (ev Ascending) Swap(i, j int)      { ev[i], ev[j] = ev[j], ev[i] }

// Descending sorts a slice of events in reverse chronological order (newest
// first)
type Descending []*E

func (e Descending) Len() int           { return len(e) }
func (e Descending) Less(i, j int) bool { return e[i].CreatedAt.I64() > e[j].CreatedAt.I64() }
func (e Descending) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
