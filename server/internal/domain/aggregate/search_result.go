package aggregate

type SearchResult struct {
	Code      int
	Relevance float64
}

func (s SearchResult) Compare(s1 SearchResult) int {
	if s.Relevance > s1.Relevance {
		return 1
	}
	if s.Relevance < s1.Relevance {
		return -1
	}
	return 0
}
