package evaluate

// OutcomeCandidates takes condition results in template priority order. It
// excludes false conditions and records which earlier result shadows a candidate.
func OutcomeCandidates(results []Result) (eligible []int, blocked map[int]int) {
	blocked = make(map[int]int)
	for i, result := range results {
		if result.State == False {
			continue
		}
		if len(eligible) > 0 && results[eligible[len(eligible)-1]].State == True {
			blocked[i] = eligible[len(eligible)-1]
			continue
		}
		shadowed := false
		for _, earlier := range eligible {
			if result.Implies(results[earlier]) {
				blocked[i] = earlier
				shadowed = true
				break
			}
		}
		if !shadowed {
			eligible = append(eligible, i)
		}
	}
	return eligible, blocked
}
