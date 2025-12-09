# Wikispecies Name Parser

This package parses Wikispecies formatted taxonomic names using a PEG (Parsing Expression Grammar) parser.

## Usage

```go
import "github.com/sfborg/harvester/internal/sources/wikisp/wsparser"

result, err := wsparser.Parse("''Felis catus'' Linnaeus, 1758")
if err != nil {
    // handle error
}

fmt.Println(result.Canonical) // "Felis catus"
fmt.Println(result.Author)    // "Linnaeus"
fmt.Println(result.Year)      // "1758"

// With template containing full and short author names
result, _ = wsparser.Parse("''Canis lupus'' {{a|Carl Linnaeus|L.}}, 1758")
fmt.Println(result.Author)      // "Carl Linnaeus"
fmt.Println(result.ShortAuthor) // "L."

// With bracketed author containing target and display text
result, _ = wsparser.Parse("''Felis catus'' [[Carl Linnaeus|Linnaeus]], 1758")
fmt.Println(result.Author)      // "Carl Linnaeus"
fmt.Println(result.ShortAuthor) // "Linnaeus"

// With reference
result, _ = wsparser.Parse("''Felis catus'' Linnaeus, 1758: Systema Naturae")
fmt.Println(result.Reference) // ": Systema Naturae"

// With unparsed tail (no colon)
result, _ = wsparser.Parse("''Canis lupus'' extra text")
fmt.Println(result.Tail)      // " extra text"
```

## How It Works

The parser uses **position capture during parsing** rather than AST walking:

1. **Grammar Definition** (`name.peg`): Defines the PEG grammar with embedded action blocks `{ }` that capture start/end positions as tokens are matched

2. **Position Storage** (`wikisp_name.go`): The `WikispName` struct stores integer positions for each component (canonical name, author, year, tail)

3. **Substring Extraction** (`wikisp_name.go`): After parsing completes, `Extract()` uses the captured positions to extract actual substrings from the original input

4. **Public API** (`parse.go`): `Parse()` function orchestrates the entire process and returns `ParsedName`

### Why Position Capture vs AST Walking?

**Position capture** (current approach):
- ✅ More efficient - no tree traversal needed
- ✅ Simpler code - direct substring extraction
- ✅ Lower memory overhead - only stores positions
- ✅ Idiomatic for PEG tools - leverages action blocks

**AST walking** (alternative):
- ❌ Requires traversing the parse tree after parsing
- ❌ More complex code with visitor pattern
- ❌ Higher memory usage - stores full tree structure
- ✅ More flexible if you need tree transformations

## Supported Formats

The grammar handles these Wikispecies markup patterns:

- **Italic names**: `''Homo sapiens''`
- **Bold italic**: `'''Homo sapiens'''`
- **With author**: `''Felis catus'' Linnaeus`
- **With year**: `''Felis catus'' Linnaeus, 1758`
- **Bracketed author**: `[[Carl Linnaeus|Linnaeus]]` (captures both target and display)
- **Template author**: `{{a|Carl Linnaeus|L.}}` (captures both full and short form)
- **With reference**: `''Homo sapiens'': 123` (everything after `:`)
- **With tail**: `''Canis lupus'' unparsed text` (trailing text without `:`)
- **Templates in names**: `''{{BASEPAGENAME}}''`

**Not supported**: Bare names without italic markup (Wikispecies typically uses markup for scientific names)

### Author vs ShortAuthor

The parser captures both full and abbreviated author names when available:

- **Author** (`result.Author`): Full author name
  - From templates: `{{a|Carl Linnaeus|L.}}` → `"Carl Linnaeus"`
  - From wikilinks: `[[Carl Linnaeus|Linnaeus]]` → `"Carl Linnaeus"` (target)
  - Plain text: `Linnaeus` → `"Linnaeus"`

- **ShortAuthor** (`result.ShortAuthor`): Abbreviated form used in scientific citations
  - From templates: `{{a|Carl Linnaeus|L.}}` → `"L."`
  - From wikilinks: `[[Carl Linnaeus|Linnaeus]]` → `"Linnaeus"` (display text)
  - Plain text: (empty, no short form available)

This gives you flexibility: use `ShortAuthor` for scientific name citations (e.g., "*Homo sapiens* L.") while keeping the full name for reference.

### Reference vs Tail

The parser distinguishes between two types of trailing content:

- **Reference** (`result.Reference`): Everything after a `:` separator, typically bibliographic references (e.g., `: Systema Naturae, 10: 123`)
- **Tail** (`result.Tail`): Unparsed trailing text when there's no `:`, such as comments or other markup that doesn't fit the expected pattern

Only one of these will be populated - if there's a `:`, it captures as Reference; otherwise any trailing text is captured as Tail.

## Modifying the Grammar

1. Edit `name.peg` to change the grammar
2. Run `just peg` to regenerate the parser
3. Update tests in `parse_test.go`
4. Run `go test` to verify

### PEG Syntax Notes

- Use `!` for negation, not regex `^` (e.g., `(!"}}" .)+` not `[^\}]+`)
- Action blocks `{ }` contain Go code executed during parsing
- `token.begin` and `token.end` provide character positions
- Ordered choice `/` tries alternatives left-to-right
- `?` = optional, `*` = zero or more, `+` = one or more

## Parse Quality

The parser assigns a quality score (0-3) to each parsed name:

- **0 (Unparseable)**: Parse failed completely
- **1 (Partial)**: Parsed canonical name but has unparsed tail content
- **2 (Good)**: Parsed canonical name cleanly (may have reference)
- **3 (Complete)**: Fully parsed with canonical, author, and year

```go
result, _ := wsparser.Parse("''Homo sapiens'' Linnaeus, 1758")
fmt.Println(result.Quality) // 3 (Complete)

result, _ = wsparser.Parse("''Homo sapiens'' unparsed text")
fmt.Println(result.Quality) // 1 (Partial)
```

Use quality scores to:
- Track parser improvement over time
- Identify problematic input patterns
- Filter results for downstream processing

## Testing with Real Data

The parser is tested against real Wikispecies data using a JSON Lines test file format:

### Test Data Format

Each line in `testdata/wikisp_names.jsonl` is a JSON object:

```json
{"in":"''Homo sapiens'' {{a|Carolus Linnaeus|Linnaeus}}, 1758: {{BHL|page/6981790|20}}","out":{"canonical":"Homo sapiens","author":"Carolus Linnaeus","shortAuthor":"Linnaeus","year":"1758","reference":": {{BHL|page/6981790|20}}"}}
```

### Generating Test Data

Use the `testgen` tool to populate expected output from parser results:

```bash
# Generate test data from input file
go run ./internal/sources/wikisp/wsparser/testgen \
  testdata/wikisp_names.txt \
  testdata/wikisp_names.jsonl

# Review the output file and manually fix any incorrect results
# Then run tests
go test ./internal/sources/wikisp/wsparser -run TestWikispeciesNames

# Show quality statistics
go run ./internal/sources/wikisp/wsparser/testgen stats testdata/wikisp_names.jsonl
```

Output:
```
Quality Distribution (10 total):
================================
  0 (Unparseable):      3 ( 30.0%)
  1 (Partial):          5 ( 50.0%)
  2 (Good):             1 ( 10.0%)
  3 (Complete):         1 ( 10.0%)
```

**Workflow:**
1. Add new test inputs to `testdata/wikisp_names.txt` (one per line)
2. Run `testgen` to generate expected outputs with quality scores
3. Use `testgen stats` to see quality distribution
4. Review `testdata/wikisp_names.jsonl` and correct any wrong outputs
5. Run tests to validate

This approach scales to millions of test cases without making test files unwieldy.

## Files

- `name.peg` - Grammar definition with action blocks
- `name.peg.go` - Generated parser (don't edit manually)
- `wikisp_name.go` - Position storage and extraction
- `parse.go` - Public API
- `parse_test.go` - Unit tests
- `testdata_test.go` - Real-world data tests
- `example_test.go` - Usage examples
- `testgen/main.go` - Tool to generate test data from parser output
