# WFWP Implementation Plan

## Overview
Integration of World Ferns and World Plants (WFWP) data processing into
the harvester architecture. The original wfwp code is ~1320 lines with
complex taxonomy hierarchy processing.

## Current State
- **Original code**: `/home/dimus/code/golang/wfwp/main.go` (persist_id
  branch) - **USE THIS AS REFERENCE IMPLEMENTATION**
- **Target location**: `internal/sources/worldplants/`
- **Label**: `wfwp` (already configured in list.go)
- **Original code stats**: ~1320 lines, single file implementation
- **Branch**: `persist_id` (confirmed active)

### Implementation Progress
- ✅ **Phase 1**: File Structure Setup (wp.go)
- ✅ **Phase 2**: Core Types and Utilities (helpers.go)
- ✅ **Phase 3**: Name Parsing (parse.go + parse_test.go)
- ✅ **Phase 4**: CSV Processing and Hierarchy (hierarchy.go)
- ✅ **Phase 5**: Synonym Processing (synonym.go)
- ✅ **Phase 6**: References (integrated into synonym.go)
- ✅ **Phase 7**: Metadata (meta.go)
- ✅ **Phase 8**: Main ToSfga Implementation (sfga.go)
- ✅ **Phase 9**: File Preparation - Extract Override (extract.go)
- ⏳ **Phase 10**: Configuration and Testing (pending)

**Next Phase:** Phase 10 - Testing and final integration!

## Data Input Requirements

**IMPORTANT**: WFWP data is NOT in public domain and must be provided by
the user.

### Input Format
- User provides a directory path via `--dir` CLI option
- Directory must contain ONLY wfwp CSV files
- Two types of CSV files are recognized:

#### 1. Ferns Data
- **Filename**: `ferns.csv`
- **Content**: World Ferns taxonomic data
- **Processing**: Used as-is

#### 2. Plants Data (Multiple Files)
- **Filename pattern**: `/\d+\.csv/` (e.g., `1.csv`, `2.csv`, `10.csv`)
- **Content**: World Plants taxonomic data split across files
- **Processing**: Must be:
  1. Sorted numerically by filename (1, 2, 3... not 1, 10, 2)
  2. Concatenated into single `plants.csv`
  3. Headers removed from all files except the first

### Usage Example
```bash
harvester get wfwp --dir=/path/to/wfwp-data/
```

Where `/path/to/wfwp-data/` contains:
```
ferns.csv
1.csv
2.csv
3.csv
...
50.csv
```

### Processing Output
After concatenation:
- `ferns.csv` → processed as dataset 1140
- `plants.csv` (concatenated) → processed as dataset 1141

## Original Code Reference Map

Use the original implementation at `/home/dimus/code/golang/wfwp/main.go`
(persist_id branch) as the authoritative reference. Key sections:

| Function/Section | Lines | Purpose |
|-----------------|-------|---------|
| `rank()` | 26-74 | Rank string to coldp.Rank mapping |
| `rankBySuffix()` | 76-112 | Infer rank from taxonomic suffix |
| `rankLevel()` | 114-135 | Rank order for hierarchy comparison |
| `formatAuthors()` | 212-228 | Format author lists |
| `wfwpParse()` | 230-343 | Main name parser with gnparser |
| `getPersistentId()` | 345-368 | Generate stable IDs from hierarchy |
| `getNode()` | 370-420 | Parse CSV line into hNode |
| `handleInfraspecies()` | 422-437 | Handle trinomial parents |
| `isAutonym()` | 439-450 | Detect autonyms |
| `spFromAutonym()` | 452-476 | Create species from autonym |
| `rankGroup()` | 478-489 | Group ranks (species/genus/higher) |
| `synonymRank()` | 491-539 | Infer synonym rank |
| `addReference()` | 541-583 | Parse and add references |
| `handleMetadata()` | 585-696 | Fetch metadata from API |
| `cleanYear()` | 705-723 | Validate publication year |
| `getBasionymId()` | 725-729 | Create basionym lookup key |
| `getParentId()` | 740-748 | Find parent in rankStack |
| `ImportWPWF()` | 750-1240 | **Main processing loop** |
| `concatenateWPFiles()` | 1242-1299 | Concatenate numbered CSV files |
| `main()` | 1301-1320 | Entry point |

**Critical algorithms to preserve:**
1. **Persistent ID generation** (345-368): Must use EXACT same algorithm
2. **Hierarchy building** (750-911): rankStack management is complex
3. **Synonym processing** (998-1135): Basionym linking logic
4. **Autonym detection** (877-886): Creates implicit species records

## Architecture Analysis

### WFWP Data Flow in Harvester
```
0. Prepare input files
   - User provides directory via --dir option
   - Concatenate numbered CSV files into plants.csv
   - Validate ferns.csv exists
   ↓
1. Read CSV file (pipe-delimited)
   - Process ferns.csv (dataset 1140)
   - Process plants.csv (dataset 1141)
   ↓
2. Parse each line → create hNode
   ↓
3. Build hierarchy using rankStack
   ↓
4. Handle special cases (autonyms, infraspecies)
   ↓
5. Generate persistent IDs based on classification path
   ↓
6. Process synonyms for each accepted name
   ↓
7. Handle references, distributions, vernaculars
   ↓
8. Insert into SFGA database
   ↓
9. Fetch and insert metadata from ChecklistBank API
```

### CSV Format (pipe-delimited)
```
rank|num|accepted|citation|vernacular|distribution|synonyms|status|remarks
```

### Key Components in Original Code

#### 1. Data Structures (~80 lines)
- `gnparsed` - parsed name components
- `classification` - taxonomic hierarchy
- `hNode` - hierarchical node with metadata
- `citation` - reference information

#### 2. Rank Handling (~90 lines)
- `rank()` - convert string to coldp.Rank
- `rankBySuffix()` - infer rank from taxonomic suffix
- `rankLevel()` - get rank order for comparison
- `rankGroup()` - group ranks (species/genus/higher)

#### 3. Name Parsing (~115 lines)
- `wfwpParse()` - main parser wrapper for gnparser
- `formatAuthors()` - format author lists
- Handles: manuscript names, hybrids, combinations

#### 4. Hierarchy Building (~120 lines)
- `getNode()` - parse CSV line into hNode
- `handleInfraspecies()` - handle trinomials
- `isAutonym()` - detect autonyms
- `spFromAutonym()` - create species from autonym
- `getParentId()` - find parent in rankStack
- `getPersistentId()` - generate stable IDs

#### 5. Synonym Processing (~170 lines in main loop)
- Parse synonym list (= delimited)
- Handle obsolete ranks
- Fix hybrid notation
- Detect duplicate synonyms
- Link basionyms

#### 6. Reference Handling (~90 lines)
- `addReference()` - parse citations
- `getBasionymId()` - create basionym lookup key
- `cleanYear()` - validate publication year

#### 7. Metadata (~110 lines)
- `handleMetadata()` - fetch from ChecklistBank API
- Parse YAML response
- Add contributors

#### 8. Main Processing Loop (~470 lines)
- Read CSV file
- Build nodes array with rankStack
- Generate persistent IDs
- Process each node:
  - Create NameUsage
  - Process synonyms
  - Handle basionyms
  - Add distributions
  - Add vernaculars
- Batch insert to database

## Implementation Plan

### Phase 1: File Structure Setup ✓
**Deliverable:** Create initial project structure
- [✓] `wp.go` - main struct and New()

**Files roadmap** (created in subsequent phases):
- [✓] `helpers.go` - types, rank functions, utility (Phase 2)
- [✓] `parse.go` - name parsing logic (Phase 3)
- [✓] `parse_test.go` - parsing tests (Phase 3)
- [✓] `extract.go` - file preparation (Phase 9)
- [✓] `hierarchy.go` - node and hierarchy building (Phase 4)
- [✓] `synonym.go` - synonym processing (Phase 5, includes reference handling)
- [✓] `meta.go` - metadata fetching (Phase 7)
- [✓] `sfga.go` - ToSfga() implementation (Phase 8)
- [ ] `wfwp_test.go` - integration tests (Phase 10)

### Phase 2: Core Types and Utilities ✓
**File: helpers.go**
- [✓] Define hNode struct
- [✓] Define citation struct
- [✓] Implement rank() function
- [✓] Implement rankBySuffix()
- [✓] Implement rankLevel()
- [✓] Implement rankGroup()
- [✓] Implement getParentId()
- [✓] Implement getPersistentId()
- [ ] Add helper tests (deferred)

### Phase 3: Name Parsing ✓
**File: parse.go**
- [✓] Define gnparsed struct
- [✓] Implement formatAuthors()
- [✓] Implement wfwpParse() with:
  - [✓] Manuscript name handling
  - [✓] Hybrid detection
  - [✓] Authorship extraction
  - [✓] Rank inference
- [✓] Implement isAutonym()
- [✓] Implement spFromAutonym()
- [✓] Implement synonymRank()
- [✓] Implement getBasionymId()

**File: parse_test.go**
- [✓] Test formatAuthors
- [✓] Test simple binomials
- [✓] Test trinomials
- [✓] Test manuscript names
- [✓] Test combinations
- [✓] Test hybrids
- [✓] Test autonym detection

### Phase 4: CSV Processing and Hierarchy ✓
**File: hierarchy.go**
- [✓] Implement getNode() - parse CSV line
- [✓] Implement buildHierarchy():
  - [✓] Read CSV file
  - [✓] Build rankStack
  - [✓] Handle order/family/genus ranks
  - [✓] Handle species rank
  - [✓] Handle infraspecies ranks
  - [✓] Detect and create autonyms
  - [✓] Skip illegal names
- [✓] Add validation for:
  - [✓] Skipped names (incertae sedis, etc.)
  - [✓] Quadrinomials
  - [✓] Clades and subgroups

**Implementation notes:**
- Broke into 9 focused functions for readability
- `getNode()` - parses pipe-delimited CSV
- `validateName()` - filters illegal names
- `buildHierarchy()` - main orchestrator, reads CSV and builds structure
- `processNodeByRank()` - routes to appropriate handler
- 5 rank-specific handlers (Order, Species, Infraspecies, Higher, Default)
- Autonym handling returns optional parent node to insert
- Progress logging every 1000 rows
- Returns both node slice and map for lookups

### Phase 5: Synonym Processing ✓
**File: synonym.go**
- [✓] Implement processSynonyms():
  - [✓] Parse synonym list (= delimiter)
  - [✓] Filter obsolete ranks
  - [✓] Fix hybrid notation (xGenus → × Genus)
  - [✓] Parse synonym with gnparser
  - [✓] Create synonym NameUsage
  - [✓] Handle duplicate detection
- [✓] Implement basionym handling:
  - [✓] Build basionym lookup
  - [✓] Link basionyms to combinations
  - [✓] Handle ambiguous basionyms (blacklist)
- [ ] Add tests for edge cases (deferred)

**Implementation notes:**
- Broke into 11 focused functions for clarity
- `processSynonyms()` - main orchestrator
- `parseSynonymList()` - splits by = delimiter
- `shouldSkipSynonym()` - filters obsolete ranks
- `parseSynonymString()` - extracts name and reference
- `fixHybridNotation()` - fixes xGenus → × Genus
- `isValidSynonym()` - validates quality and cardinality
- `createSynonymUsage()` - creates coldp.NameUsage
- `updateBasionymLookup()` - tracks basionyms
- `linkBasionyms()` - links combinations to basionyms
- `addReference()` - adds citations (also used for accepted names)
- `cleanYear()` - validates publication year
- Includes duplicate detection and ambiguous basionym blacklisting

### Phase 6: References ✓
**File: synonym.go** (integrated with synonym processing)
- [✓] Implement addReference():
  - [✓] Parse citation string
  - [✓] Extract year and page
  - [✓] Build full citation
  - [✓] Generate reference ID
  - [✓] Deduplicate references
- [✓] Implement cleanYear()
- [✓] citation struct defined in helpers.go
- [ ] Add reference tests (deferred)

**Implementation notes:**
- Reference handling integrated into synonym.go
- `addReference()` is used by both accepted names and synonyms
- `cleanYear()` validates 4-digit years (1700-2099)
- Automatic deduplication via lookup map

### Phase 7: Metadata ✓
**File: meta.go**
- [✓] Implement fetchMetadata():
  - [✓] HTTP GET from ChecklistBank API
  - [✓] Parse YAML response
  - [✓] Extract creators and contributors
  - [✓] Build coldp.Meta struct
- [✓] Add hardcoded contributors (Dima, Geoff)
- [✓] Handle missing metadata gracefully
- [✓] Support both datasets (ferns=1140, plants=1141)

**Implementation notes:**
- Broke into 8 focused functions for clarity
- `fetchMetadata()` - main entry point, fetches from ChecklistBank API
- `buildMetaFromResponse()` - constructs coldp.Meta
- `extractKey()` - converts key to string
- `extractString()` - safe string extraction
- `extractInt()` - safe int extraction
- `extractAlias()` - builds alias with suffix
- `extractCreators()` - parses creator list, identifies contact
- `extractStringFromMap()` - helper for map extraction
- `getHardcodedContributors()` - returns Dima and Geoff
- Fetches from: https://api.checklistbank.org/dataset/{datasetID}.yaml
- Identifies Michael Hassler as contact person
- Graceful handling of missing/malformed data

### Phase 8: Main ToSfga Implementation ✓
**File: sfga.go**

**SFGA Schema Context** ([schema.sql](https://github.com/sfborg/sfga/blob/main/schema.sql)):
- Uses NAME, TAXON, SYNONYM, REFERENCE, VERNACULAR, DISTRIBUTION, METADATA tables
- sflib provides: InsertReferences, InsertNameUsages, InsertDistributions, InsertVernaculars, InsertMeta
- Insertion order: References → NameUsages → Distributions → Vernaculars → Metadata

**IMPORTANT**: WFWP generates TWO separate SFGA files (ferns and plants),
not one.

- [✓] Implement ToSfga() orchestrator:
  - [✓] Process ferns dataset (1140)
  - [✓] Process plants dataset (1141)

- [✓] Implement processDataset():
  1. [✓] Build node hierarchy
  2. [✓] Generate persistent IDs
  3. [✓] Process each node (create accepted NameUsage)
  4. [✓] Process synonyms and link basionyms
  5. [✓] Add references, distributions, vernaculars
  6. [✓] Insert to SFGA in correct order
  7. [✓] Fetch and insert metadata

**Implementation notes:**
- Broke into 8 focused functions for clarity
- `ToSfga()` - main orchestrator, processes both datasets
- `processDataset()` - processes single dataset (ferns or plants)
- `generatePersistentIDs()` - replaces temp UUIDs with stable IDs
- `processAllNodes()` - main processing loop
- `createAcceptedNameUsage()` - creates accepted name records
- `processVernaculars()` - handles common names
- `insertToSfga()` - inserts in correct order
- `datasetRecords` - struct to hold all records
- Handles dual SFGA output (ferns.sfga, plants.sfga)
- Progress logging every 1000 nodes
- Proper error handling and propagation
- Database connection lifecycle (connect, insert, close)

### Phase 9: File Preparation (Extract Override) ✓
**File: extract.go**

This phase prepares the input files from the user-provided directory.

- [✓] Implement concatenatePlants():
  - [✓] List all files in input directory
  - [✓] Find files matching `/^\d+\.csv$/` pattern
  - [✓] Sort files numerically (natural sort):
    - Use strconv.Atoi() to extract numbers
    - Sort: [1.csv, 2.csv, 10.csv] NOT [1.csv, 10.csv, 2.csv]
  - [✓] Create plants.csv in extract directory:
    - [✓] Read first file completely (with header)
    - [✓] For subsequent files, skip first line (header)
    - [✓] Stream/buffer to handle large files
    - [✓] Use bytes operations for efficient processing
  - [✓] Log progress (e.g., "Concatenated 50 files into plants.csv")

- [✓] Implement validateInputDir():
  - [✓] Check input directory exists and is readable
  - [✓] Verify ferns.csv exists
  - [✓] Verify at least one numbered CSV file exists
  - [✓] Return error with helpful message if validation fails

- [✓] Override Extract() method:
  - [✓] Signature: Extract(path string) error (matches interface)
  - [✓] Call validateInputDir() on input directory
  - [✓] Call concatenatePlants() to create plants.csv
  - [✓] Copy ferns.csv to extract directory
  - [✓] Log file sizes for verification

**Implementation notes:**
- Broke into 15 small, focused functions for readability
- Each function has single responsibility
- Error messages are helpful and specific
- Progress logging every 10 files during concatenation

### Phase 10: Configuration and Testing
- [ ] Add config options if needed:
  - [ ] Dataset ID (1140 for ferns, 1141 for plants)
  - [ ] Suffix for metadata
  - [ ] Verbose ID mode (for debugging)
- [ ] Create test data (small.csv from wfwp)
- [ ] Write integration tests
- [ ] Test with harvester CLI:
  - [ ] `harvester list`
  - [ ] `harvester get wfwp --dir=/path/to/data/`

## Dependencies
- ✓ gnparser (botanical code, with details)
- ✓ sflib (SFGA, CoLDP types)
- ✓ google/uuid (for persistent IDs)
- yaml.v3 (for metadata parsing)
- http client (for ChecklistBank API)

## Testing Strategy

### Unit Tests
1. **Parse tests** - name parsing, authorship, hybrids
2. **Rank tests** - rank conversion, inference, levels
3. **Hierarchy tests** - parent finding, ID generation
4. **Synonym tests** - parsing, deduplication
5. **Reference tests** - citation parsing, year cleaning

### Integration Tests
1. **Small dataset** - use wfwp/data/small.csv
2. **Full ferns** - if available
3. **Full plants** - if available

### Edge Cases to Test
- Autonyms (Abies alba subsp. alba)
- Hybrid taxa (× notation)
- Manuscript names (ined., comb. ined.)
- Obsolete ranks (convar., race, etc.)
- Quadrinomials (should be excluded)
- Duplicate names
- Missing parents
- Circular references

## File Organization Strategy

The original code (`/home/dimus/code/golang/wfwp/main.go`) is a single
~1320 line file. We're modularizing it for better maintainability while
preserving exact behavior:

| Original Lines | Target File | Status | Notes |
|---------------|-------------|---------|-------|
| 26-155 | `helpers.go` | ✓ Complete | rank functions, IDs |
| 148-343 | `parse.go` | ✓ Complete | name parsing, types |
| 148-343 | `parse_test.go` | ✓ Complete | parsing tests |
| 370-476, 838-911 | `hierarchy.go` | ✓ Complete | CSV reading, node building |
| 998-1135, 541-583, 705-729 | `synonym.go` | ✓ Complete | synonym + reference handling |
| 585-696 | `meta.go` | ✓ Complete | metadata fetching |
| 750-1240 | `sfga.go` | ✓ Complete | main ToSfga loop |
| 1242-1299 | `extract.go` | ✓ Complete | file concatenation |
| N/A | `wfwp_test.go` | Pending | integration tests |

**Modularization principles:**
- Keep function signatures identical where possible
- Preserve exact algorithms (especially ID generation)
- Add struct methods instead of free functions where appropriate
- Tests should verify behavior matches original

## Migration Notes

### Changes from Original Code
1. **Removed global variables** - `badNodes` map (line 210)
2. **Split into modules** - better separation of concerns (see table above)
3. **Error handling** - return errors instead of `os.Exit()`
4. **Config integration** - use harvester config, not env vars
5. **Logging** - use slog instead of `fmt.Println`
6. **Testing** - comprehensive test coverage
7. **Parser setup** - initialized in `worldplants` struct, not in functions

### Preserved Behavior
1. **Persistent ID generation** - same algorithm
2. **Namespace** - "SFBORG::WFWP"
3. **Synonym processing** - same rules
4. **Rank handling** - same mappings
5. **Metadata** - same ChecklistBank integration

## Timeline Estimate

| Phase | Estimated Time | Dependencies |
|-------|---------------|--------------|
| Phase 1 | ✓ Done | None |
| Phase 2 | 2 hours | Phase 1 |
| Phase 3 | 3 hours | Phase 2 |
| Phase 4 | 4 hours | Phase 2, 3 |
| Phase 5 | 3 hours | Phase 3, 4 |
| Phase 6 | 2 hours | Phase 2 |
| Phase 7 | 2 hours | None |
| Phase 8 | 4 hours | Phase 2-7 |
| Phase 9 | 1 hour | Phase 8 |
| Phase 10 | 3 hours | Phase 8, 9 |
| **Total** | **24 hours** | |

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Complex hierarchy logic | High | Thorough testing, incremental dev |
| Persistent ID consistency | High | Use exact same algorithm |
| Memory usage (large files) | Medium | Batch processing, streaming |
| API availability | Low | Graceful fallback, caching |
| CSV format changes | Medium | Validation, clear error messages |

## Success Criteria

1. ✓ Dataset appears in `harvester list` as "wfwp"
2. [ ] File preparation works correctly:
   - [ ] Numbered CSV files concatenated in correct order
   - [ ] Headers handled properly (kept in first, removed in rest)
   - [ ] Both ferns.csv and plants.csv available for processing
3. [ ] Dual output generation works:
   - [ ] Two separate SFGA files created (ferns and plants)
   - [ ] Filename logic works for both `{}` and non-`{}` cases
   - [ ] Each SFGA contains correct dataset
4. [ ] Can process small.csv without errors
5. [ ] Persistent IDs match original wfwp output
6. [ ] All tests pass
7. [ ] Can process full ferns dataset (dataset 1140)
8. [ ] Can process full plants dataset (dataset 1141)
9. [ ] Metadata fetched and inserted correctly for both datasets
10. [ ] Both outputs validate as proper SFGA format
11. [ ] CLI usage works:
    - [ ] `harvester get wfwp --dir=/path/to/data/`
    - [ ] With output pattern: `harvester get wfwp --dir=/path/to/data/
          -o output_{}.sfga`

## Next Steps

1. Review and approve this plan
2. Start Phase 2 (Core Types and Utilities)
3. Implement incrementally with tests
4. Review after each phase
5. Integration testing at Phase 10
