package helpers

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func NormalizeString(value string) string {
	t := transform.Chain(
		norm.NFD,
		runes.Remove(runes.In(unicode.Mn)), // Elimina marcas diacríticas
		norm.NFC,
	)

	result, _, _ := transform.String(t, value)

	// Convertir a minúsculas
	result = strings.ToLower(result)

	// Reemplazar espacios por guiones
	result = strings.ReplaceAll(result, " ", "-")
	result = strings.ReplaceAll(result, "_", "-")

	// Opcional: eliminar caracteres que no sean alfanuméricos o guiones
	result = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' || r == '.' {
			return r
		}
		return -1
	}, result)

	return result
}

func Toggle(value, fst, snd string) string {
	if value == fst {
		return snd
	}
	return fst
}
