# Harvester

Harvester converts biodiversity datasets from various formats into [SFGA]
(Species File Group Archive) — a standardized format for taxonomic data.

## Introduction

Quite often, important biodiversity datasets do not have a standard archive
package and cannot be easily converted to one by using the [SF] project.
Their data may be distributed as custom CSV/TSV dumps, proprietary databases,
or complex archives with no common structure.

Harvester bridges this gap. It downloads datasets from their original sources,
extracts and parses the data, and transforms it into [SFGA] — a uniform
SQLite-based format covering taxonomic names, synonyms, vernacular names,
references, and distributions. Currently supported sources include GRIN, ITIS,
NCBI, ION, IOC Birds, PaleoDB, Wikispecies, and World Plants.
SFGA files can by easily converted to other popular biodiversity formats using
[SF] application.

## Installation

### Binary release

Download the latest binary for your platform from the [releases page].

### From source

Requires Go 1.25+.

```bash
go install github.com/sfborg/harvester@latest
```

## Usage

### List available data sources

```bash
harvester list        # show all supported datasets
harvester list -v     # show details for each dataset
```

### Convert a dataset

```bash
harvester get <label> <output-file> # download and convert to SFGA
harvester get <label> -s <output>   # skip download, use cached data
harvester get <label> -z <output>   # output as compressed zip file
```

Replace `<label>` with a dataset identifier or its row number from
`harvester list`.
For the output, provide only the file name. Several files will be
generated.

### Example

```bash
harvester list
harvester list -v
harvester get itis ~/tmp/itis
harvester get 3 ~/tmp/itis
harvester get ncbi -z ~/tmp/ncbi
```

If no output target is given, converted files are saved as SFGA archives in the
current directory.

## Output format

Harvester produces [SFGA] archives — SQLite
databases following a common schema for taxonomic names, synonyms, vernacular
names, references, and distributions.

| Extension       | Description                        |
|-----------------|------------------------------------|
| `.sql`          | SQL dump (plain text)              |
| `.sqlite`       | SQLite binary database             |
| `.sql.zip`      | SQL dump, compressed               |
| `.sqlite.zip`   | SQLite binary database, compressed |

Note that when providing output path only the file name is needed, extensions
will be added automatically.


## Development

See [AGENTS.md] for build instructions, architecture overview,
and how to add new data sources.

## Authors

* [Dmitry Mozzherin]
* [Geoffrey Ower]

## License

Released under [MIT license]

## Artificial Intelligence Policy

We use artificial intelligence to help find algorithms, decide on
implementation approaches, and generate code. All automatically generated
code is carefully reviewed, with inconsistencies fixed, superfluous
implementations removed, and optimizations improved. No code that we do
not understand or approve makes it into published versions of Harvester. We
primarily use Claude Code, with limited use of Gemini CLI.

[AGENTS.md]: AGENTS.md
[Dmitry Mozzherin]: https://github.com/dimus
[Geoffrey Ower]: https://github.com/gdower
[MIT license]: LICENSE
[SFGA]: https://github.com/sfborg/sfga
[SF]: https://github.com/sfborg/sf
[releases page]: https://github.com/sfborg/harvester/releases/latest
