package worldplants

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/sfborg/sflib/pkg/coldp"
	"gopkg.in/yaml.v3"
)

// fetchMetadata retrieves metadata from ChecklistBank API and returns
// a coldp.Meta struct.
// Reference: original lines 585-696
func (wp *worldplants) fetchMetadata(
	datasetID string,
	issuedDate string,
	version string,
	suffix string,
) (*coldp.Meta, error) {
	slog.Info("Fetching metadata", "datasetID", datasetID)

	url := fmt.Sprintf(
		"https://api.checklistbank.org/dataset/%s.yaml",
		datasetID,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"metadata fetch returned status %d",
			resp.StatusCode,
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata response: %w", err)
	}

	metadata := make(map[string]any)
	err = yaml.Unmarshal(body, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	meta, err := wp.buildMetaFromResponse(
		metadata,
		issuedDate,
		version,
		suffix,
	)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

// buildMetaFromResponse constructs a coldp.Meta from the API response.
func (wp *worldplants) buildMetaFromResponse(
	metadata map[string]any,
	issuedDate string,
	version string,
	suffix string,
) (*coldp.Meta, error) {
	key := extractKey(metadata)
	title := extractString(metadata, "title")
	alias := extractAlias(metadata, suffix)
	description := extractString(metadata, "description")
	geographicScope := extractString(metadata, "geographicScope")
	taxonomicScope := extractString(metadata, "taxonomicScope")
	confidence := extractInt(metadata, "confidence")
	completeness := extractInt(metadata, "completeness")
	license := extractString(metadata, "license")
	url := extractString(metadata, "url")
	logo := extractString(metadata, "logo")

	creators, contact := extractCreators(metadata)
	contributors := getHardcodedContributors()

	meta := &coldp.Meta{
		Key:             key,
		Title:           title,
		Alias:           alias,
		Description:     description,
		DOI:             "",
		Issued:          issuedDate,
		Version:         version,
		GeographicScope: geographicScope,
		TaxonomicScope:  taxonomicScope,
		Confidence:      confidence,
		Completeness:    completeness,
		License:         license,
		URL:             url,
		Logo:            logo,
		Contact:         contact,
		Creators:        creators,
		Contributors:    contributors,
	}

	return meta, nil
}

// extractKey extracts the dataset key from metadata.
func extractKey(metadata map[string]any) string {
	if keyValue, ok := metadata["key"]; ok {
		if keyInt, ok := keyValue.(int); ok {
			return strconv.Itoa(keyInt)
		}
	}
	return ""
}

// extractString safely extracts a string value from metadata.
func extractString(metadata map[string]any, key string) string {
	if value, ok := metadata[key]; ok {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return ""
}

// extractInt safely extracts an int value from metadata.
func extractInt(metadata map[string]any, key string) int {
	if value, ok := metadata[key]; ok {
		if intValue, ok := value.(int); ok {
			return intValue
		}
	}
	return 0
}

// extractAlias builds the alias with suffix.
func extractAlias(metadata map[string]any, suffix string) string {
	alias := extractString(metadata, "alias")
	if suffix != "" {
		return strings.TrimSpace(alias + " " + suffix)
	}
	return alias
}

// extractCreators extracts creator list and identifies contact person.
func extractCreators(metadata map[string]any) ([]coldp.Actor, *coldp.Actor) {
	var creators []coldp.Actor
	var contact *coldp.Actor

	creatorList, ok := metadata["creator"]
	if !ok {
		return creators, nil
	}

	creatorSlice, ok := creatorList.([]interface{})
	if !ok {
		return creators, nil
	}

	for _, creator := range creatorSlice {
		creatorMap, ok := creator.(map[string]interface{})
		if !ok {
			continue
		}

		actor := coldp.Actor{
			Given:        extractStringFromMap(creatorMap, "given"),
			Family:       extractStringFromMap(creatorMap, "family"),
			Email:        extractStringFromMap(creatorMap, "email"),
			City:         extractStringFromMap(creatorMap, "city"),
			Country:      extractStringFromMap(creatorMap, "country"),
			Organization: extractStringFromMap(creatorMap, "organisation"),
		}

		// Identify contact person (Michael Hassler)
		if actor.Email == "hassler.michael@t-online.de" {
			contactCopy := actor
			contact = &contactCopy
		}

		creators = append(creators, actor)
	}

	return creators, contact
}

// extractStringFromMap safely extracts a string from a map.
func extractStringFromMap(m map[string]interface{}, key string) string {
	if value, ok := m[key]; ok {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return ""
}

// getHardcodedContributors returns the hardcoded contributors
// (Dima and Geoff).
func getHardcodedContributors() []coldp.Actor {
	dima := coldp.Actor{
		Given:        "Dmitry",
		Family:       "Mozzherin",
		Email:        "dmozzherin@gmail.com",
		Orcid:        "0000-0003-1593-1417",
		RorID:        "047426m28",
		City:         "Champaign",
		State:        "Illinois",
		Country:      "US",
		Organization: "Illinois Natural History Survey",
		Note:         "COL Pipeline Developer",
	}

	geoff := coldp.Actor{
		Given:        "Geoffrey",
		Family:       "Ower",
		Email:        "gdower@illinois.edu",
		Orcid:        "0000-0002-9770-2345",
		RorID:        "047426m28",
		City:         "Champaign",
		State:        "Illinois",
		Country:      "US",
		Organization: "Illinois Natural History Survey",
		Note:         "COL Pipeline Developer",
	}

	return []coldp.Actor{dima, geoff}
}
