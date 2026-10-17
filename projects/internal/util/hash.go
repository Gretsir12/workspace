package util

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

func HashDependencies(deps []string) string {
	parts := make([]string, 0, len(deps))
	for _, d := range deps {
		// нормализуйте имя/версию (trim, lower-case и т.п.), если нужно
		parts = append(parts, string(d))
	}
	sort.Strings(parts)

	sum := sha256.Sum256([]byte(strings.Join(parts, "\n")))

	return hex.EncodeToString(sum[:])
}
