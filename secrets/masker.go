package secrets

import (
	"sort"
	"unicode"
)

type SecretMasker struct {
	values    []string
	generator []func(string) string
}

var DefaultMasker = NewSecretMasker()

func NewSecretMasker() *SecretMasker {
	return &SecretMasker{
		values:    []string{},
		generator: []func(string) string{},
	}
}

func (s *SecretMasker) AddGenerator(gen func(string) string) {
	s.generator = append(s.generator, gen)
}

func (s *SecretMasker) AddValue(value string) {
	if value == "" {
		return
	}
	s.values = append(s.values, value)

	for _, gen := range s.generator {
		value = gen(value)
		s.values = append(s.values, value)
	}

	sort.Strings(s.values)
	reverseStrings(s.values)
}

func reverseStrings(a []string) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

func (s *SecretMasker) ApplyGenerators(input string) string {
	for _, gen := range s.generator {
		input = gen(input)
	}
	return input
}

func (s *SecretMasker) Mask(input string) string {
	if len(input) == 0 || len(s.values) == 0 {
		return input
	}

	hits := searchAll(input, s.values)
	if len(hits) == 0 {
		return input
	}

	input = replace(input, hits, "****")

	return input
}

type searchHit struct {
	Start  int
	Length int
}

func (s *searchHit) End() int {
	return s.Start + s.Length
}

func (s *searchHit) Empty() bool {
	return s.Start < 0 || s.Length <= 0
}

func search(haystack []rune, needle []rune) []searchHit {
	if len(needle) == 0 || len(haystack) == 0 {
		return []searchHit{}
	}

	l := len(haystack)
	n := len(needle)
	if l < n {
		return []searchHit{}
	}

	j := 0

	hits := []searchHit{}
	for i := 0; i < l; i++ {
		if j == n {
			hit := searchHit{Start: i - j, Length: n}
			hits = append(hits, hit)
			j = 0
		}

		if j < n {
			sr := haystack[i]
			tr := needle[j]

			if sr == tr {
				j++
				continue
			}

			if tr < sr {
				tr, sr = sr, tr
			}

			if 'A' <= sr && sr <= 'Z' && tr == sr+'a'-'A' {
				j++
				continue
			}

			r := unicode.SimpleFold(sr)
			for r != sr && r < tr {
				r = unicode.SimpleFold(r)
			}
			if r == tr {
				j++
				continue
			}
		}

		j = 0
	}

	return hits
}

func searchAll(haystack string, needles []string) []searchHit {
	allHits := []searchHit{}
	for _, needle := range needles {
		hits := search([]rune(haystack), []rune(needle))

		if len(hits) == 0 {
			continue
		}

		for _, hit := range hits {
			match := false
			replaceIdx := -1
			for i, existingHit := range allHits {
				if existingHit.Start == hit.Start {
					match = true
					if hit.Length > existingHit.Length {
						replaceIdx = i
					}

					break
				}

				if hit.Start > existingHit.Start && hit.Start < existingHit.End() {
					// skip overlapping hits
					match = true
				}

				if hit.Start < existingHit.Start && hit.End() > existingHit.Start {
					// skip overlapping hits
					match = true
					replaceIdx = i
				}
			}
			if !match {
				allHits = append(allHits, hit)
			} else if replaceIdx != -1 {
				allHits[replaceIdx] = hit
			}
		}
	}

	sort.Slice(allHits, func(i, j int) bool {
		return allHits[i].Start < allHits[j].Start
	})

	return allHits
}

func replace(haystack string, hits []searchHit, replacement string) string {
	if len(hits) == 0 {
		return haystack
	}

	var result []rune
	startIndex := 0
	target := []rune(haystack)
	r := []rune(replacement)

	for _, hit := range hits {
		length := hit.Start - startIndex
		precedingText := target[startIndex : startIndex+length]

		if len(precedingText) > 0 {
			result = append(result, precedingText...)
		}
		result = append(result, r...)
		startIndex = hit.End()
	}

	remainingTest := target[startIndex:]
	if len(remainingTest) > 0 {
		result = append(result, remainingTest...)
	}

	return string(result)
}
