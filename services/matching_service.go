package services

import (
	"strings"

	"lost-found-school/models"
)

// Bobot: Nama 25, Kategori 20, Warna 15, Merek 15, Ciri 15, Lokasi 5, Tanggal 5
func ScoreReports(lost, found models.Report) float64 {
	score := 0.0
	if containsWord(lost.ItemName, found.ItemName) || containsWord(found.ItemName, lost.ItemName) {
		score += 25
	} else if strings.EqualFold(strings.TrimSpace(lost.ItemName), strings.TrimSpace(found.ItemName)) {
		score += 25
	} else if similarity(lost.ItemName, found.ItemName) > 0.5 {
		score += 15
	}
	if lost.CategoryID != 0 && lost.CategoryID == found.CategoryID {
		score += 20
	}
	if eqFold(lost.Color, found.Color) && lost.Color != "" {
		score += 15
	}
	if eqFold(lost.Brand, found.Brand) && lost.Brand != "" {
		score += 15
	}
	if lost.Characteristics != "" && found.Characteristics != "" {
		if similarity(lost.Characteristics, found.Characteristics) > 0.4 {
			score += 15
		} else if shareWord(lost.Characteristics, found.Characteristics) {
			score += 8
		}
	}
	if eqFold(lost.Location, found.Location) && lost.Location != "" {
		score += 5
	} else if shareWord(lost.Location, found.Location) && lost.Location != "" {
		score += 3
	}
	if lost.EventDate != "" && lost.EventDate == found.EventDate {
		score += 5
	}
	if score > 100 {
		score = 100
	}
	return score
}

func Label(score float64) string {
	switch {
	case score >= 70:
		return "Kemungkinan cocok"
	case score >= 40:
		return "Mungkin cocok"
	default:
		return "Kemungkinan rendah"
	}
}

func eqFold(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

func words(s string) []string {
	return strings.Fields(strings.ToLower(s))
}

func containsWord(hay, needle string) bool {
	hay = strings.ToLower(hay)
	for _, w := range words(needle) {
		if len(w) >= 4 && strings.Contains(hay, w) {
			return true
		}
	}
	return false
}

func shareWord(a, b string) bool {
	set := map[string]bool{}
	for _, w := range words(a) {
		if len(w) >= 4 {
			set[w] = true
		}
	}
	for _, w := range words(b) {
		if len(w) >= 4 && set[w] {
			return true
		}
	}
	return false
}

// Jaccard similarity on words
func similarity(a, b string) float64 {
	wa := words(a)
	wb := words(b)
	if len(wa) == 0 || len(wb) == 0 {
		return 0
	}
	setA := map[string]bool{}
	for _, w := range wa {
		setA[w] = true
	}
	inter := 0
	setU := map[string]bool{}
	for _, w := range wa {
		setU[w] = true
	}
	for _, w := range wb {
		setU[w] = true
		if setA[w] {
			inter++
		}
	}
	if len(setU) == 0 {
		return 0
	}
	return float64(inter) / float64(len(setU))
}
