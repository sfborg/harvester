package paleodb

import "github.com/sfborg/sflib/pkg/coldp"

// Taken from the API
// "subspecies","2" "species","3" "subgenus","4" "genus","5" "subtribe","6"
// "tribe","7" "subfamily","8" "family","9" "superfamily","10"
// "infraorder","11" "suborder","12" "order","13" "superorder","14"
// "infraclass","15" "subclass","16" "class","17" "superclass","18"
// "subphylum","19" "phylum","20" "superphylum","21" "subkingdom","22"
// "kingdom","23" "unranked clade","25" "informal","26"
var rankMap = map[string]coldp.Rank{
	"2":  coldp.Subspecies,
	"3":  coldp.Species,
	"4":  coldp.Subgenus,
	"5":  coldp.Genus,
	"6":  coldp.Subtribe,
	"7":  coldp.Tribe,
	"8":  coldp.Subfamily,
	"9":  coldp.Family,
	"10": coldp.Superfamily,
	"11": coldp.Infraorder,
	"12": coldp.Suborder,
	"13": coldp.Order,
	"14": coldp.Superorder,
	"15": coldp.Infraclass,
	"16": coldp.Subclass,
	"17": coldp.Class,
	"18": coldp.Superclass,
	"19": coldp.Subphylum,
	"20": coldp.Phylum,
	"21": coldp.Superphylum,
	"22": coldp.Subkingdom,
	"23": coldp.Kingdom,
	"24": coldp.UnknownRank,
	"25": coldp.UnknownRank,
	"26": coldp.UnknownRank,
}
