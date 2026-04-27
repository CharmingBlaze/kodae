package check

import "sort"

// levenshtein returns the edit distance between a and b.
func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for j := 0; j <= len(rb); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 0
			if ra[i-1] != rb[j-1] {
				cost = 1
			}
			del := prev[j] + 1
			ins := cur[j-1] + 1
			sub := prev[j-1] + cost
			cur[j] = min3(del, ins, sub)
		}
		prev, cur = cur, prev
	}
	return prev[len(rb)]
}

func min3(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= c {
		return b
	}
	return c
}

// maxEditOK is the largest edit distance we still treat as a likely typo for a string of length n.
func maxEditOK(n int) int {
	if n <= 1 {
		return 0
	}
	if n <= 8 {
		return 2
	}
	if n <= 20 {
		return 3
	}
	return 4
}

// suggestName picks the closest candidate to name by Levenshtein, if it is a clear best match.
func suggestName(name string, candidates []string) (string, bool) {
	seen := make(map[string]struct{}, len(candidates))
	var distinct []string
	for _, c := range candidates {
		if c == "" {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		distinct = append(distinct, c)
	}
	if len(distinct) == 0 {
		return "", false
	}
	maxD := maxEditOK(len(name))
	type match struct {
		c string
		d int
	}
	var ms []match
	for _, c := range distinct {
		if c == name {
			return c, true
		}
		d := levenshtein(name, c)
		if d <= maxD {
			ms = append(ms, match{c: c, d: d})
		}
	}
	if len(ms) == 0 {
		return "", false
	}
	sort.Slice(ms, func(i, j int) bool {
		if ms[i].d != ms[j].d {
			return ms[i].d < ms[j].d
		}
		return ms[i].c < ms[j].c
	})
	if len(ms) >= 2 && ms[0].d == ms[1].d {
		return "", false
	}
	return ms[0].c, true
}
