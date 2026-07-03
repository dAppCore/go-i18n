package i18n

import "testing"

func BenchmarkForCategory(b *testing.B) {
	msg := Message{
		Text:  "plain",
		Zero:  "zero",
		One:   "one",
		Two:   "two",
		Few:   "few",
		Many:  "many",
		Other: "other",
	}
	cats := []PluralCategory{PluralZero, PluralOne, PluralTwo, PluralFew, PluralMany, PluralOther}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, cat := range cats {
			benchServiceStringSink = msg.ForCategory(cat)
		}
	}
}

func BenchmarkByteUnitName(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = byteUnitName()
	}
}

func BenchmarkLogWordName(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchServiceStringSink = logWordName()
	}
}
