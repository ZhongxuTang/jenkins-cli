package util

func UpdateRecent(recent []string, value string, limit int) []string {
	if value == "" {
		return recent
	}
	next := make([]string, 0, limit)
	next = append(next, value)
	for _, item := range recent {
		if item == "" || item == value {
			continue
		}
		next = append(next, item)
		if limit > 0 && len(next) >= limit {
			break
		}
	}
	return next
}

func FilterRecent(recent []string, allowSet map[string]struct{}, limit int) []string {
	if len(recent) == 0 {
		return recent
	}
	filtered := make([]string, 0, len(recent))
	for _, value := range recent {
		if value == "" {
			continue
		}
		if _, ok := allowSet[value]; ok {
			filtered = append(filtered, value)
			if limit > 0 && len(filtered) >= limit {
				break
			}
		}
	}
	return filtered
}

func BuildAllowSet(items []string) map[string]struct{} {
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		set[item] = struct{}{}
	}
	return set
}
