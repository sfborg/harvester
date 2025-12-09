package wsparser_test

import (
	"testing"

	"github.com/sfborg/harvester/internal/sources/wikisp/wsparser"
)

func TestParse(t *testing.T) {
	tests := []struct {
		msg            string
		input          string
		wantName       string
		wantAuthorship string
		wantRef        string
		wantTail       string
		wantError      bool
	}{
		{
			msg:      "bare uninomial",
			input:    "Homo",
			wantName: "Homo",
		},
		{
			msg:      "bare binomial",
			input:    "Homo sapiens",
			wantName: "Homo sapiens",
		},
		{
			msg:      "bare trinomial",
			input:    "Homo sapiens sapiens",
			wantName: "Homo sapiens sapiens",
		},
		{
			msg:      "simple italic name",
			input:    "''Homo sapiens''",
			wantName: "Homo sapiens",
		},
		{
			msg:      "name with author",
			input:    "''Homo sapiens'' Linnaeus",
			wantName: "Homo sapiens",
			wantTail: " Linnaeus",
		},
		{
			msg:            "name with author and year",
			input:          "''Homo sapiens'' {{a|Linnaeus}}, 1758",
			wantName:       "Homo sapiens",
			wantAuthorship: "Linnaeus, 1758",
		},
		{
			msg:            "name with bracketed author",
			input:          "''Felis catus'' [[Carl Linnaeus|Linnaeus]], 1758",
			wantName:       "Felis catus",
			wantAuthorship: "Linnaeus, 1758",
		},
		{
			msg:            "name with template author (short form)",
			input:          "''Canis lupus'' {{a|Linnaeus|L.}}, 1758",
			wantName:       "Canis lupus",
			wantAuthorship: "L., 1758",
		},
		{
			msg:            "name with template author (short form)",
			input:          "''Canis lupus'' {{a|Linnaeus|L.}} 1758",
			wantName:       "Canis lupus",
			wantAuthorship: "L. 1758",
		},
		{
			msg:            "name with template author (no short form)",
			input:          "''Canis lupus'' {{a|Linnaeus}}, 1758",
			wantName:       "Canis lupus",
			wantAuthorship: "Linnaeus, 1758",
		},
		{
			msg:      "bold italic name",
			input:    "'''Homo sapiens sapiens'''",
			wantName: "Homo sapiens sapiens",
		},
		{
			msg:      "name with reference after colon",
			input:    "''Homo sapiens'': 123",
			wantName: "Homo sapiens",
			wantRef:  ": 123",
		},
		{
			msg:            "name with author and reference",
			input:          "''Felis catus'' [[Linnaeus]], 1758: Systema Naturae",
			wantName:       "Felis catus",
			wantAuthorship: "Linnaeus, 1758",
			wantRef:        ": Systema Naturae",
		},
		{
			msg:            "name with author and reference",
			input:          "''Felis catus'' {{aut|Linnaeus}}, 1758: Systema Naturae",
			wantName:       "Felis catus",
			wantAuthorship: "Linnaeus, 1758",
			wantRef:        ": Systema Naturae",
		},
		{
			msg:            "name with author and reference",
			input:          "''Felis catus'' {{au|Linnaeus}}, 1758: Systema Naturae",
			wantName:       "Felis catus",
			wantAuthorship: "Linnaeus, 1758",
			wantRef:        ": Systema Naturae",
		},
		{
			msg:      "name with unparsed tail (no colon)",
			input:    "''Canis lupus'' some unparsed text",
			wantName: "Canis lupus",
			wantTail: " some unparsed text",
		},
		{
			msg:            "template author with reference (short form)",
			input:          "''Passer domesticus'' {{a|Linnaeus|L.}}, 1758: 456",
			wantName:       "Passer domesticus",
			wantAuthorship: "L., 1758",
			wantRef:        ": 456",
		},
		{
			msg: "cat=t name (category=taxonomist)",
			input: "''Abacetus inexpectatus'' " +
				"{{a|Oleg Leonidovich Kryzhanovsky|Kryzhanovskij|cat=t}} & " +
				"{{a|Gayirbeg Magomedovich Abdurakhmanov|Abdurakhmanov|cat=t}}, 1983",
			wantName:       "Abacetus inexpectatus",
			wantAuthorship: "Kryzhanovskij & Abdurakhmanov, 1983",
		},
		{
			msg:      "infraspecies with rank ({{AN}} means autonym)",
			input:    "''Quercus kiukiangensis'' var. ''kiukiangensis'', {{AN}}",
			wantName: "Quercus kiukiangensis var. kiukiangensis",
			wantTail: ", {{AN}}",
		},
		{
			msg: "infraspecies with rank and species authorship {{AN}} means autonym",
			input: "''Panicum alatum'' {{a|Zuloaga}} & " +
				"{{a|Osvaldo Morrone|Morrone}} var. ''alatum'' {{AN}}",
			wantName:       "Panicum alatum var. alatum",
			wantAuthorship: "",
			wantTail:       " {{AN}}",
		},
		{
			msg:            "original authorship",
			input:          "''Coreura fida'' ([[Hübner]], 1827)",
			wantName:       "Coreura fida",
			wantAuthorship: "(Hübner, 1827)",
		},
		{
			msg:            "simple authorship template",
			input:          "Emydinae {{a|Samuel Booker McDowell, Jr.|McDowell}}, 1964",
			wantName:       "Emydinae",
			wantAuthorship: "McDowell, 1964",
		},
		{
			msg:            "simple authorship template 2",
			input:          "Carabus hungaricus scythus {{a|Victor Ivanovich Motschulsky|Motschulsky}}, 1847",
			wantName:       "Carabus hungaricus scythus",
			wantAuthorship: "Motschulsky, 1847",
		},
		{
			msg:            "hybrid",
			input:          "× ''Elyhordeum elymoides'' {{a|Xifreda}}, Kurtziana 28: 292 (2000)",
			wantName:       "× Elyhordeum elymoides",
			wantAuthorship: "Xifreda",
			wantTail:       ", Kurtziana 28: 292 (2000)",
		},
		{
			msg:            "extinct 1 with et al",
			input:          "†''Sahelanthropus tchadensis'' [[Brunet]] ''et al''., 2002",
			wantName:       "†Sahelanthropus tchadensis",
			wantAuthorship: "Brunet et al., 2002",
		},
		{
			msg:            "extinct 2",
			input:          "†Lepidodendraceae {{a|Endl.}}, Gen. Pl.: 70. 1836.",
			wantName:       "†Lepidodendraceae",
			wantAuthorship: "Endl.",
			wantTail:       ", Gen. Pl.: 70. 1836.",
		},
		{
			msg:            "year in parentheses",
			input:          "Lauderiaceae ([[Franz Schütt|Schütt]]) [[Lemmermann]] (1899)",
			wantName:       "Lauderiaceae",
			wantAuthorship: "(Schütt) Lemmermann (1899)",
		},
		{
			msg:            "year in parentheses",
			input:          "Lauderiaceae ([[Franz Schütt|Schütt]]) [[Lemmermann]] (1899)",
			wantName:       "Lauderiaceae",
			wantAuthorship: "(Schütt) Lemmermann (1899)",
		},
		{
			msg:      "wiki link in canonical name",
			input:    "''[[Testudo kleinmanni]]'' Lortet, 1883",
			wantName: "Testudo kleinmanni",
			wantTail: " Lortet, 1883",
		},
		{
			msg:      "wiki link with display text in canonical",
			input:    "''[[Brachypodium sylvaticum|Brachypodium]]'' Beauv.",
			wantName: "Brachypodium",
			wantTail: " Beauv.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			got, err := wsparser.Parse(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("Parse() error = %v, wantError %v",
					err, tt.wantError)
				return
			}
			if err != nil {
				return
			}

			if got.Canonical != tt.wantName {
				t.Errorf("Parse() Canonical = %q, want %q",
					got.Canonical, tt.wantName)
			}
			if got.Authorship != tt.wantAuthorship {
				t.Errorf("Parse() Authorship = %q, want %q",
					got.Authorship, tt.wantAuthorship)
			}
			if got.Reference != tt.wantRef {
				t.Errorf("Parse() Reference = %q, want %q",
					got.Reference, tt.wantRef)
			}
			if got.Tail != tt.wantTail {
				t.Errorf("Parse() Tail = %q, want %q",
					got.Tail, tt.wantTail)
			}
		})
	}
}
