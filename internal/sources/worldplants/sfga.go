package worldplants

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnames/gnlib/ent/nomcode"
	"github.com/sfborg/sflib/pkg/coldp"
	"github.com/sfborg/sflib/pkg/sfga"
)

// ToSfga implements the main SFGA conversion logic.
// WFWP generates TWO SFGA files (ferns and plants).
// Reference: original lines 750-1240, 1301-1320
func (wp *worldplants) ToSfga(sfgaArchive sfga.Archive) error {
	slog.Info("Starting WFWP SFGA conversion")

	extractDir := wp.cfg.ExtractDir

	// Prompt user for dataset choice
	choice := wp.promptDatasetChoice()

	// Process based on user choice
	if choice == "plants" {
		// Process plants dataset (1141)
		plantsPath := filepath.Join(extractDir, "plants.csv")
		slog.Info("Processing plants dataset", "path", plantsPath)
		err := wp.processDataset(plantsPath, "1141", "plants", sfgaArchive)
		if err != nil {
			return fmt.Errorf("failed to process plants: %w", err)
		}
	} else {
		// Process ferns dataset (1140) - default
		fernsPath := filepath.Join(extractDir, "ferns.csv")
		slog.Info("Processing ferns dataset", "path", fernsPath)
		err := wp.processDataset(fernsPath, "1140", "ferns", sfgaArchive)
		if err != nil {
			return fmt.Errorf("failed to process ferns: %w", err)
		}
	}

	slog.Info("WFWP SFGA conversion complete")
	return nil
}

// promptDatasetChoice prompts the user to choose which dataset to process.
// Returns "ferns" (default) or "plants" based on user input.
func (wp *worldplants) promptDatasetChoice() string {
	fmt.Print("Choose dataset [f=ferns (default), p=plants]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		slog.Warn("Error reading input, using default (ferns)", "error", err)
		return "ferns"
	}

	choice := strings.TrimSpace(strings.ToLower(input))

	if choice == "p" || choice == "plants" {
		return "plants"
	}

	// Default to ferns for any other input (including empty)
	return "ferns"
}

// processDataset processes a single dataset (ferns or plants).
func (wp *worldplants) processDataset(
	csvPath string,
	datasetID string,
	suffix string,
	sfgaArchive sfga.Archive,
) error {
	slog.Info("Building hierarchy", "dataset", datasetID)

	// Build hierarchy from CSV
	nodes, nodeMap, err := wp.buildHierarchy(csvPath)
	if err != nil {
		return fmt.Errorf("failed to build hierarchy: %w", err)
	}

	slog.Info("Generating persistent IDs", "nodes", len(nodes))

	// Generate persistent IDs
	persistentIDs := wp.generatePersistentIDs(nodes, nodeMap)

	slog.Info("Processing nodes to create records")

	// Process all nodes to create records
	records, err := wp.processAllNodes(
		nodes,
		persistentIDs,
		nodeMap,
	)
	if err != nil {
		return fmt.Errorf("failed to process nodes: %w", err)
	}

	slog.Info(
		"Inserting to SFGA",
		"references", len(records.references),
		"nameUsages", len(records.nameUsages),
		"distributions", len(records.distributions),
		"vernaculars", len(records.vernaculars),
	)

	// Connect to database
	_, err = sfgaArchive.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to SFGA: %w", err)
	}
	defer sfgaArchive.Close()

	// Insert to SFGA
	err = wp.insertToSfga(sfgaArchive, records)
	if err != nil {
		return fmt.Errorf("failed to insert to SFGA: %w", err)
	}

	slog.Info("Fetching and inserting metadata", "dataset", datasetID)

	// Fetch and insert metadata
	meta, err := wp.fetchMetadata(
		datasetID,
		wp.cfg.ArchiveDate,
		wp.cfg.ArchiveVersion,
		suffix,
	)
	if err != nil {
		return fmt.Errorf("failed to fetch metadata: %w", err)
	}

	err = sfgaArchive.InsertMeta(meta)
	if err != nil {
		return fmt.Errorf("failed to insert metadata: %w", err)
	}

	return nil
}

// datasetRecords holds all records for a dataset.
type datasetRecords struct {
	references    []coldp.Reference
	nameUsages    []coldp.NameUsage
	distributions []coldp.Distribution
	vernaculars   []coldp.Vernacular
}

// generatePersistentIDs generates persistent IDs for all nodes.
// Reference: original lines 913-939
func (wp *worldplants) generatePersistentIDs(
	nodes []hNode,
	nodeMap map[string]hNode,
) map[string]string {
	persistentIDs := make(map[string]string)
	uniqueIDs := make(map[string]struct{})

	for _, node := range nodes {
		// Generate persistent ID for this node
		if _, exists := persistentIDs[node.id]; !exists {
			persistentID := getPersistentId(node, nodeMap, wp.namespace, false)

			// Check for duplicates
			if _, duplicate := uniqueIDs[persistentID]; duplicate {
				slog.Warn(
					"Duplicate persistent ID (skipping)",
					"name", node.verbatimName,
					"id", persistentID,
				)
				continue
			}

			uniqueIDs[persistentID] = struct{}{}
			persistentIDs[node.id] = persistentID
		}

		// Generate persistent ID for parent if needed
		if node.parentId != "" {
			if _, exists := persistentIDs[node.parentId]; !exists {
				parentNode := nodeMap[node.parentId]
				persistentID := getPersistentId(
					parentNode,
					nodeMap,
					wp.namespace,
					false,
				)
				persistentIDs[node.parentId] = persistentID
			}
		}
	}

	return persistentIDs
}

// processAllNodes processes all nodes to create SFGA records.
// Reference: original lines 913-1165
func (wp *worldplants) processAllNodes(
	nodes []hNode,
	persistentIDs map[string]string,
	nodeMap map[string]hNode,
) (*datasetRecords, error) {
	records := &datasetRecords{
		references:    []coldp.Reference{},
		nameUsages:    []coldp.NameUsage{},
		distributions: []coldp.Distribution{},
		vernaculars:   []coldp.Vernacular{},
	}

	referenceLookup := make(map[string]citation)
	allBasionyms := make(basionymLookup)

	// Track name usages for basionym linking
	var allNameUsages []*coldp.NameUsage

	for i, node := range nodes {
		if (i+1)%1000 == 0 {
			slog.Info("Processing node", "count", i+1, "total", len(nodes))
		}

		persistentID := persistentIDs[node.id]
		if persistentID == "" {
			continue // Skipped duplicate
		}

		// Create accepted name usage
		acceptedUsage, err := wp.createAcceptedNameUsage(
			node,
			persistentID,
			persistentIDs,
			referenceLookup,
		)
		if err != nil {
			slog.Warn("Failed to create accepted name", "error", err)
			continue
		}

		records.nameUsages = append(records.nameUsages, *acceptedUsage)
		allNameUsages = append(allNameUsages, acceptedUsage)

		// Process synonyms
		synonyms, basionyms, err := wp.processSynonyms(
			node,
			persistentID,
			referenceLookup,
		)
		if err != nil {
			slog.Warn("Failed to process synonyms", "error", err)
		}

		// Add synonyms to name usages
		for _, syn := range synonyms {
			records.nameUsages = append(records.nameUsages, *syn)
			allNameUsages = append(allNameUsages, syn)
		}

		// Merge basionyms
		for k, v := range basionyms {
			allBasionyms[k] = v
		}

		// Process distributions
		if node.verbatimDistribution != "" {
			dist := coldp.Distribution{
				TaxonID:   persistentID,
				Area:      strings.TrimSpace(node.verbatimDistribution),
				Gazetteer: coldp.TextGz,
			}
			records.distributions = append(records.distributions, dist)
		}

		// Process vernaculars
		if node.verbatimVernacularName != "" {
			verns := wp.processVernaculars(persistentID, node.verbatimVernacularName)
			records.vernaculars = append(records.vernaculars, verns...)
		}
	}

	slog.Info("Linking basionyms", "count", len(allBasionyms))

	// Link basionyms to combinations
	err := linkBasionyms(allNameUsages, allBasionyms, wp)
	if err != nil {
		return nil, fmt.Errorf("failed to link basionyms: %w", err)
	}

	// Convert reference lookup to slice
	for _, ref := range referenceLookup {
		records.references = append(records.references, coldp.Reference{
			ID:       ref.id,
			Author:   ref.author,
			Issued:   ref.year,
			Title:    ref.title,
			Citation: ref.citation,
		})
	}

	return records, nil
}

// createAcceptedNameUsage creates a NameUsage for an accepted name.
func (wp *worldplants) createAcceptedNameUsage(
	node hNode,
	persistentID string,
	persistentIDs map[string]string,
	referenceLookup map[string]citation,
) (*coldp.NameUsage, error) {
	// Get parent ID
	parentID := ""
	if node.parentId != "" {
		parentID = persistentIDs[node.parentId]
		// Nil UUID means no parent
		if parentID == "00000000-0000-0000-0000-000000000000" {
			parentID = ""
		}
	}

	// Add reference
	refID, page, year := addReference(
		referenceLookup,
		node.verbatimCitation,
		node.parsed,
		wp.namespace,
	)

	// Generate link for species-group ranks
	link := ""
	if node.rank == coldp.Species ||
		node.rank == coldp.Subspecies ||
		node.rank == coldp.Variety ||
		node.rank == coldp.Form {
		link = "http://www.worldplants.de/?deeplink=" +
			strings.ReplaceAll(node.parsed.canonicalFull, " ", "-")
	}

	scientificNameString := node.parsed.canonicalFull + " " +
		node.parsed.authorship

	usage := &coldp.NameUsage{
		ID:                   persistentID,
		ParentID:             parentID,
		ScientificName:       node.parsed.canonicalFull,
		Rank:                 node.rank,
		Uninomial:            node.parsed.uninomial,
		GenericName:          node.parsed.genus,
		InfragenericEpithet:  node.parsed.subgenus,
		SpecificEpithet:      node.parsed.species,
		InfraspecificEpithet: node.parsed.infraspecies,
		Notho:                node.parsed.notho,
		TaxonomicStatus:      coldp.AcceptedTS,
		Authorship:           node.parsed.authorship,
		NamePhrase:           node.parsed.appendedPhrase,
		NameStatus:           node.parsed.nameStatus,
		ReferenceID:          refID,
		NameReferenceID:      refID,
		PublishedInPage:      page,
		PublishedInYear:      cleanYear(year),
		Code:                 nomcode.Botanical,
		Link:                 link,
		Remarks:              node.parsed.remarks,
		ScientificNameString: scientificNameString,
	}

	return usage, nil
}

// processVernaculars processes vernacular names.
func (wp *worldplants) processVernaculars(
	taxonID string,
	verbatimVernaculars string,
) []coldp.Vernacular {
	var result []coldp.Vernacular

	names := strings.Split(verbatimVernaculars, ",")
	for _, name := range names {
		vern := coldp.Vernacular{
			TaxonID: taxonID,
			Name:    strings.TrimSpace(name),
		}
		result = append(result, vern)
	}

	return result
}

// insertToSfga inserts all records to SFGA in the correct order.
func (wp *worldplants) insertToSfga(
	sfgaArchive sfga.Archive,
	records *datasetRecords,
) error {
	var err error

	// Insert in correct order
	slog.Info("Inserting references", "count", len(records.references))
	err = sfgaArchive.InsertReferences(records.references)
	if err != nil {
		return fmt.Errorf("failed to insert references: %w", err)
	}

	// Deduplicate name usages (original lines 1203-1214)
	uniqueUsages := make(map[string]coldp.NameUsage)
	var deduplicatedUsages []coldp.NameUsage
	for _, usage := range records.nameUsages {
		if _, exists := uniqueUsages[usage.ID]; exists {
			slog.Warn("Duplicate name usage found (skipping)", "id", usage.ID)
			continue
		}
		uniqueUsages[usage.ID] = usage
		deduplicatedUsages = append(deduplicatedUsages, usage)
	}

	slog.Info(
		"Inserting name usages",
		"count", len(deduplicatedUsages),
		"duplicates_removed", len(records.nameUsages)-len(deduplicatedUsages),
	)
	err = sfgaArchive.InsertNameUsages(deduplicatedUsages)
	if err != nil {
		return fmt.Errorf("failed to insert name usages: %w", err)
	}

	slog.Info("Inserting distributions", "count", len(records.distributions))
	err = sfgaArchive.InsertDistributions(records.distributions)
	if err != nil {
		return fmt.Errorf("failed to insert distributions: %w", err)
	}

	slog.Info("Inserting vernaculars", "count", len(records.vernaculars))
	err = sfgaArchive.InsertVernaculars(records.vernaculars)
	if err != nil {
		return fmt.Errorf("failed to insert vernaculars: %w", err)
	}

	return nil
}
