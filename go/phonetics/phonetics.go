// Package phonetics gives the grammar engine an ear. It vendors the CMU
// Pronouncing Dictionary (~134k words as ARPABET phonemes with stress
// digits) and answers the questions spelling cannot: does a word START with
// a vowel SOUND (ewe does not, hour does), is its final syllable STRESSED
// (commit is, visit is not), how many syllables, what does it rhyme with.
//
// Base English here is the English of England, per house policy: a small
// override table re-hears the words where the American dictionary and
// en-GB part ways (herb keeps its /h/). The en-US locale re-hears those
// words back through its own locale data, which outranks this package.
//
//	vowel, known := phonetics.StartsWithVowelSound("ewe")   // false, true — "a ewe"
//	stressed, _ := phonetics.FinalSyllableStressed("commit") // true — "committed"
package phonetics

import (
	// Note: AX-6 — the vendored dictionary ships gzip-compressed; pinned core
	// has no decompression primitive.
	"compress/gzip"
	// Note: AX-6 — go:embed requires the embed package for the dictionary asset.
	_ "embed"
	// Note: AX-6 — gzip.NewReader wants an io.Reader over the embedded bytes;
	// pinned core has no bytes-reader primitive.
	"bytes"
	// Note: AX-6 — io.ReadAll drains the decompressed stream.
	"io"
	// Note: AX-6 — the dictionary parse happens exactly once, lazily.
	"sync"

	"dappco.re/go"
)

//go:embed cmudict.dict.gz
var cmudictGz []byte

// gbOverrides re-hears headwords where the English of England disagrees
// with the American source dictionary. Values are ARPABET phoneme strings
// in the same shape as dictionary lines. Kept deliberately tiny: only
// words whose divergence changes a decision this package serves.
var gbOverrides = map[string]string{
	"herb":  "HH ER1 B",
	"herbs": "HH ER1 B Z",
}

var (
	dictOnce sync.Once
	dict     map[string][]Pronunciation
)

// Lookup returns the known pronunciations of a word, primary first. The
// word is matched case-insensitively; punctuation is not stripped, so
// hyphenated headwords the dictionary knows (x-ray) resolve directly.
//
//	prons, ok := phonetics.Lookup("tomato") // two pronunciations, ok=true
func Lookup(word string) ([]Pronunciation, bool) {
	dictOnce.Do(loadDict)
	key := core.Lower(core.Trim(word))
	if key == "" {
		return nil, false
	}
	prons, ok := dict[key]
	return prons, ok
}

// Primary returns the first-listed pronunciation of a word.
//
//	pron, ok := phonetics.Primary("quiz") // [K W IH1 Z], true
func Primary(word string) (Pronunciation, bool) {
	prons, ok := Lookup(word)
	if !ok || len(prons) == 0 {
		return nil, false
	}
	return prons[0], true
}

// Size reports how many headwords the dictionary carries once loaded.
//
//	n := phonetics.Size() // > 130000
func Size() int {
	dictOnce.Do(loadDict)
	return len(dict)
}

func loadDict() {
	reader, err := gzip.NewReader(bytes.NewReader(cmudictGz))
	if err != nil {
		dict = map[string][]Pronunciation{}
		return
	}
	raw, err := io.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		dict = map[string][]Pronunciation{}
		return
	}
	lines := core.Split(string(raw), "\n")
	parsed := make(map[string][]Pronunciation, len(lines))
	for _, line := range lines {
		word, pron, ok := parseDictLine(line)
		if !ok {
			continue
		}
		parsed[word] = append(parsed[word], pron)
	}
	for word, phones := range gbOverrides {
		override := Pronunciation(core.Split(phones, " "))
		parsed[word] = append([]Pronunciation{override}, parsed[word]...)
	}
	dict = parsed
}

// parseDictLine reads one dictionary line: "word PH PH PH", with variant
// headwords as "word(2)" and optional trailing "# comment" stripped.
func parseDictLine(line string) (string, Pronunciation, bool) {
	line = core.Trim(line)
	if line == "" || core.HasPrefix(line, ";;;") || core.HasPrefix(line, "#") {
		return "", nil, false
	}
	if idx := core.Index(line, " #"); idx >= 0 {
		line = core.Trim(line[:idx])
	}
	fields := core.Fields(line)
	if len(fields) < 2 {
		return "", nil, false
	}
	word := fields[0]
	if idx := core.Index(word, "("); idx > 0 {
		word = word[:idx]
	}
	return core.Lower(word), Pronunciation(fields[1:]), true
}
