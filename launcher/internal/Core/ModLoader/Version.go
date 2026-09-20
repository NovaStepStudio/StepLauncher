package modloader

import (
	"sort"
	"strconv"
	"strings"
)

// compareTokens compara dos segmentos de una versión: si ambos son numéricos
// se comparan por valor numérico; si uno es numérico y el otro no, el numérico
// es mayor; si ninguno lo es, se comparan como texto (alpha/beta/...).
func compareTokens(a, b string) int {
	an, aerr := strconv.Atoi(a)
	bn, berr := strconv.Atoi(b)
	if aerr == nil && berr == nil {
		switch {
		case an < bn:
			return -1
		case an > bn:
			return 1
		}
		return 0
	}
	if aerr == nil {
		return 1
	}
	if berr == nil {
		return -1
	}
	return strings.Compare(a, b)
}

// CompareLoaderVersions compara dos versiones de modloader segmento a segmento
// (separa por '.' y '-') para que el orden sea numérico real y no
// lexicográfico: "21.1.248" > "21.1.99", "52.1.16" > "52.1.9".
func CompareLoaderVersions(a, b string) int {
	split := func(v string) []string {
		return strings.FieldsFunc(v, func(r rune) bool { return r == '.' || r == '-' })
	}
	ta, tb := split(a), split(b)
	n := len(ta)
	if len(tb) < n {
		n = len(tb)
	}
	for i := 0; i < n; i++ {
		if c := compareTokens(ta[i], tb[i]); c != 0 {
			return c
		}
	}
	// Prefijo idéntico: la versión con más segmentos es la más nueva.
	switch {
	case len(ta) < len(tb):
		return -1
	case len(ta) > len(tb):
		return 1
	}
	return 0
}

// SortLoaderVersionsDesc ordena las versiones de modloader de mayor a menor.
func SortLoaderVersionsDesc(versions []LoaderVersion) {
	sort.Slice(versions, func(i, j int) bool {
		return CompareLoaderVersions(versions[i].LoaderVersion, versions[j].LoaderVersion) > 0
	})
}