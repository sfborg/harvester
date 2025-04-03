# Fields meanings

## Taxonomic records

"orig_no": Original ID: This seem to be the "correct" ID we should use.
"taxon_no": Taxon ID currently accepted as most correct.
"record_type": Record type, indicating the type of data ("txn" for taxon).
"flags": Mostly empty. It might contain I= ichnotaxon. F= form taxon.
"taxon_rank": Taxon rank ID.
"taxon_name": The scientific name of the taxon.
"taxon_attr": Taxon atribution, authorship
"common_name": Common name, the vernacular name of the taxon.
"difference": For junior synonym or invalid names, gives the reason why.
"accepted_no": Accepted ID of the currently accepted name (for synonyms).
"accepted_rank": Accepted rank, the rank of the accepted name.
"accepted_name": Accepted name, the currently accepted name.
"parent_no": Parent ID
"parent_name": Parent name of its senior synonym
"immpar_no": Immediate parent ID, even if it is a junior synonym
"immpar_name": Immediate parent name, even if it is a junior synonym

### Taxonomic reference

"ref_author": Reference author.
"ref_pubyr": Reference publication year, the year the publication was released.
"reference_no": Reference ID, a unique identifier for the publication.

### Geological Taxon Information

"is_extant": Is extant, or extinct.
"n_occs": Number of fossil occurrences recorded for the taxon.
"firstapp_max_ma": First appearance maximum age (Ma),
    the maximum age of the first appearance of the taxon in millions of years.
"firstapp_min_ma": First appearance minimum age (Ma),
    the minimum age of the first appearance.
"lastapp_max_ma": Last appearance maximum age (Ma),
    the maximum age of the last appearance of the taxon.
"lastapp_min_ma": Last appearance minimum age (Ma),
    the minimum age of the last appearance.
"early_interval": Early interval, the geological time interval
    of the first appearance.
"late_interval": Late interval, the geological time interval
    of the last appearance.

### Taxon Classification

"phylum", "phylum_no", "class", "class_no", "order", "order_no",
"family", "family_no", "genus", "genus_no", "subgenus_no"

: Taxonomic classification, with the names and numbers of the phylum,
    class, order, family, genus, and subgenus.

### Type of a Taxon

"type_taxon": The name of a type taxon, if known.
"type_taxon_no": Type taxon ID.

### Taxon facts

"taxon_environment": Taxon environment, the environment in which the taxon lived.
"environment_basis": Environment basis, how the environment was determined.
"motility": Mobility, how the organism moved.
"life_habit": Life habit, the lifestyle of the organism.
"vision": Vision, information about the organism's vision.
"diet": Diet, the organism's feeding habits.
"reproduction": Reproduction, information about the organism's reproduction.
"ontogeny": Ontogeny, information about the organism's development.
"ecospace_comments": Ecospace comments, additional notes about ecological niche.
"composition": Composition, the material the organism's body was made of.
"architecture": Architecture, the organism's body structure.
"thickness": Thickness, the thickness of the organism's body or shell.
"reinforcement": Reinforcement, any structural reinforcement in the organism.
"image_no": Image number, a reference to an image of the taxon.
"primary_reference": Primary reference, the full citation of the primary publication.
"authorizer_no", "enterer_no", "modifier_no", "updater_no"
    : Internal user numbers for the database.
"authorizer", "enterer", "modifier", "updater": User names for the database.
"created": Created, the date and time the record was created.
"modified": Modified, the date and time the record was last modified.
"updated": Updated, the date and time of the last update.
