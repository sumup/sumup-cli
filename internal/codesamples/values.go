package codesamples

import (
	"math"
	"strconv"
	"strings"
)

func flattenExample(value map[string]any) map[string]any {
	result := make(map[string]any)
	var walk func(prefix string, current any)
	walk = func(prefix string, current any) {
		if prefix != "" {
			result[prefix] = current
		}
		object, ok := current.(map[string]any)
		if !ok {
			return
		}
		for key, nested := range object {
			path := key
			if prefix != "" {
				path = prefix + "." + key
			}
			walk(path, nested)
		}
	}
	walk("", value)
	return result
}

func lookupExample(flag string, values map[string]any) (any, bool) {
	if flag == "amount" {
		if value, ok := values["amount"]; ok {
			return value, true
		}
		if amount, ok := majorAmount(values); ok {
			return amount, true
		}
	}
	if flag == "tip-amount" {
		return convertedTipAmount(values)
	}
	if len(values) == 0 {
		return nil, false
	}
	paths := make([][]string, 0, len(values))
	for path := range values {
		paths = append(paths, strings.Split(path, "."))
	}
	matched, ok := bestPath(flag, paths)
	if !ok {
		return nil, false
	}
	return values[strings.Join(matched, ".")], true
}

func bestPath(flag string, paths [][]string) ([]string, bool) {
	want := strings.ReplaceAll(flag, "-", "_")
	wantPlural := want
	if !strings.HasSuffix(want, "s") {
		wantPlural = want + "s"
	}

	var best []string
	bestScore := 0
	for _, path := range paths {
		for _, candidate := range pathCandidates(path) {
			score := 0
			switch candidate {
			case want:
				score = 300 + 10 - len(path)
			case wantPlural:
				score = 200 + 10 - len(path)
			default:
				if strings.HasSuffix(candidate, "_"+want) || strings.HasSuffix(candidate, "_"+wantPlural) {
					score = 100 + 10 - len(path)
				}
			}
			if score > bestScore {
				bestScore = score
				best = path
			}
		}
	}
	if bestScore == 0 {
		return nil, false
	}
	return best, true
}

func pathCandidates(path []string) []string {
	joined := strings.Join(path, "_")
	candidates := []string{joined}
	for i := 1; i < len(path); i++ {
		candidates = append(candidates, strings.Join(path[i:], "_"))
	}
	if len(path) > 0 {
		candidates = append(candidates, path[len(path)-1])
	}
	return candidates
}

func majorAmount(values map[string]any) (string, bool) {
	total, ok := values["total_amount"].(map[string]any)
	if !ok {
		return "", false
	}
	major, ok := convertMinorUnits(total["value"], minorUnit(values, total))
	if !ok {
		return "", false
	}
	return major, true
}

func convertedTipAmount(values map[string]any) (any, bool) {
	raw, ok := values["tip_amount"]
	if !ok {
		return nil, false
	}
	total, _ := values["total_amount"].(map[string]any)
	major, ok := convertMinorUnits(raw, minorUnit(values, total))
	if !ok {
		return raw, true
	}
	return major, true
}

func minorUnit(values map[string]any, total map[string]any) int {
	for _, raw := range []any{values["minor_unit"], values["total_amount.minor_unit"]} {
		if unit, ok := intValue(raw); ok {
			return unit
		}
	}
	if total != nil {
		if unit, ok := intValue(total["minor_unit"]); ok {
			return unit
		}
	}
	return 2
}

func convertMinorUnits(raw any, unit int) (string, bool) {
	amount, ok := floatValue(raw)
	if !ok {
		return "", false
	}
	if unit < 0 {
		unit = 0
	}
	major := amount / math.Pow10(unit)
	return strconv.FormatFloat(major, 'f', -1, 64), true
}

func floatValue(raw any) (float64, bool) {
	switch value := raw.(type) {
	case float64:
		return value, true
	case int:
		return float64(value), true
	case int64:
		return float64(value), true
	default:
		return 0, false
	}
}

func intValue(raw any) (int, bool) {
	switch value := raw.(type) {
	case float64:
		return int(value), true
	case int:
		return value, true
	case int64:
		return int(value), true
	default:
		return 0, false
	}
}

func formatExampleValue(value any) []string {
	switch typed := value.(type) {
	case nil:
		return nil
	case bool:
		if !typed {
			return nil
		}
		return []string{""}
	case map[string]any:
		if enabled, ok := typed["enabled"].(bool); ok && enabled {
			return []string{""}
		}
		return nil
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			values = append(values, formatExampleValue(item)...)
		}
		return values
	case float64:
		if typed == float64(int64(typed)) {
			return []string{strconv.FormatInt(int64(typed), 10)}
		}
		return []string{strconv.FormatFloat(typed, 'f', -1, 64)}
	case string:
		if typed == "" {
			return nil
		}
		return []string{typed}
	default:
		text := strings.TrimSpace(stringify(typed))
		if text == "" {
			return nil
		}
		return []string{text}
	}
}

func stringify(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(typed)
	default:
		return ""
	}
}
