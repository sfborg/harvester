-- Prune ITIS database to ~50 records per kingdom for testing
-- Tables used by harvester:
--   taxonomic_units, hierarchy, taxon_authors_lkp, taxon_unit_types,
--   synonym_links, vernaculars, geographic_div, publications, kingdoms, version

-- Create temp table with selected TSNs (50 per kingdom, valid/accepted only)
CREATE TEMP TABLE selected_tsns AS
WITH ranked AS (
  SELECT tsn, kingdom_id, name_usage,
         ROW_NUMBER() OVER (PARTITION BY kingdom_id ORDER BY tsn) as rn
  FROM taxonomic_units
  WHERE name_usage IN ('valid', 'accepted')
    AND (unaccept_reason IS NULL OR unaccept_reason = '')
)
SELECT tsn FROM ranked WHERE rn <= 50;

-- Add some extinct taxa (from extinct.tsv list) to test extinct marking
INSERT OR IGNORE INTO selected_tsns
SELECT tsn FROM taxonomic_units
WHERE tsn IN (
  2351, 2644, 2715, 15379, 18026, 20147, 20202, 20210,
  21263, 21384, 21955, 21969, 21978, 22409, 24005, 27144,
  83088, 83089, 83090, 202423, 552479, 552505
)
AND name_usage IN ('valid', 'accepted');

-- Add all ancestor TSNs to ensure hierarchy integrity (recursive)
-- ITIS hierarchy can be 15+ levels deep, so we iterate until no new parents
INSERT OR IGNORE INTO selected_tsns
SELECT DISTINCT h.parent_tsn
FROM hierarchy h
WHERE h.tsn IN (SELECT tsn FROM selected_tsns)
  AND h.parent_tsn > 0;

-- Repeat until no new ancestors (up to 20 iterations for safety)
INSERT OR IGNORE INTO selected_tsns SELECT DISTINCT h.parent_tsn FROM hierarchy h WHERE h.tsn IN (SELECT tsn FROM selected_tsns) AND h.parent_tsn > 0;
INSERT OR IGNORE INTO selected_tsns SELECT DISTINCT h.parent_tsn FROM hierarchy h WHERE h.tsn IN (SELECT tsn FROM selected_tsns) AND h.parent_tsn > 0;
INSERT OR IGNORE INTO selected_tsns SELECT DISTINCT h.parent_tsn FROM hierarchy h WHERE h.tsn IN (SELECT tsn FROM selected_tsns) AND h.parent_tsn > 0;
INSERT OR IGNORE INTO selected_tsns SELECT DISTINCT h.parent_tsn FROM hierarchy h WHERE h.tsn IN (SELECT tsn FROM selected_tsns) AND h.parent_tsn > 0;
INSERT OR IGNORE INTO selected_tsns SELECT DISTINCT h.parent_tsn FROM hierarchy h WHERE h.tsn IN (SELECT tsn FROM selected_tsns) AND h.parent_tsn > 0;
INSERT OR IGNORE INTO selected_tsns SELECT DISTINCT h.parent_tsn FROM hierarchy h WHERE h.tsn IN (SELECT tsn FROM selected_tsns) AND h.parent_tsn > 0;
INSERT OR IGNORE INTO selected_tsns SELECT DISTINCT h.parent_tsn FROM hierarchy h WHERE h.tsn IN (SELECT tsn FROM selected_tsns) AND h.parent_tsn > 0;
INSERT OR IGNORE INTO selected_tsns SELECT DISTINCT h.parent_tsn FROM hierarchy h WHERE h.tsn IN (SELECT tsn FROM selected_tsns) AND h.parent_tsn > 0;
INSERT OR IGNORE INTO selected_tsns SELECT DISTINCT h.parent_tsn FROM hierarchy h WHERE h.tsn IN (SELECT tsn FROM selected_tsns) AND h.parent_tsn > 0;
INSERT OR IGNORE INTO selected_tsns SELECT DISTINCT h.parent_tsn FROM hierarchy h WHERE h.tsn IN (SELECT tsn FROM selected_tsns) AND h.parent_tsn > 0;
INSERT OR IGNORE INTO selected_tsns SELECT DISTINCT h.parent_tsn FROM hierarchy h WHERE h.tsn IN (SELECT tsn FROM selected_tsns) AND h.parent_tsn > 0;
INSERT OR IGNORE INTO selected_tsns SELECT DISTINCT h.parent_tsn FROM hierarchy h WHERE h.tsn IN (SELECT tsn FROM selected_tsns) AND h.parent_tsn > 0;
INSERT OR IGNORE INTO selected_tsns SELECT DISTINCT h.parent_tsn FROM hierarchy h WHERE h.tsn IN (SELECT tsn FROM selected_tsns) AND h.parent_tsn > 0;
INSERT OR IGNORE INTO selected_tsns SELECT DISTINCT h.parent_tsn FROM hierarchy h WHERE h.tsn IN (SELECT tsn FROM selected_tsns) AND h.parent_tsn > 0;
INSERT OR IGNORE INTO selected_tsns SELECT DISTINCT h.parent_tsn FROM hierarchy h WHERE h.tsn IN (SELECT tsn FROM selected_tsns) AND h.parent_tsn > 0;

-- Add some synonyms for selected taxa (excluding database artifacts)
CREATE TEMP TABLE selected_synonyms AS
SELECT sl.tsn
FROM synonym_links sl
INNER JOIN taxonomic_units tu ON sl.tsn = tu.tsn
WHERE sl.tsn_accepted IN (SELECT tsn FROM selected_tsns)
  AND COALESCE(tu.unaccept_reason, '') NOT IN (
    'unavailable, database artifact', 'database artifact'
  )
LIMIT 100;

INSERT OR IGNORE INTO selected_tsns
SELECT tsn FROM selected_synonyms;

-- Now prune tables

-- Delete from taxonomic_units (keep selected + their parents)
DELETE FROM taxonomic_units
WHERE tsn NOT IN (SELECT tsn FROM selected_tsns);

-- Delete from hierarchy
DELETE FROM hierarchy
WHERE tsn NOT IN (SELECT tsn FROM selected_tsns);

-- Delete from synonym_links
DELETE FROM synonym_links
WHERE tsn NOT IN (SELECT tsn FROM selected_tsns);

-- Delete from vernaculars (keep only those for selected taxa)
DELETE FROM vernaculars
WHERE tsn NOT IN (SELECT tsn FROM selected_tsns);

-- Delete from geographic_div
DELETE FROM geographic_div
WHERE tsn NOT IN (SELECT tsn FROM selected_tsns);

-- Get used publication IDs from reference_links
CREATE TEMP TABLE used_pubs AS
SELECT DISTINCT documentation_id as pub_id
FROM reference_links
WHERE tsn IN (SELECT tsn FROM selected_tsns)
  AND doc_id_prefix = 'PUB';

-- Delete from publications (keep only referenced ones)
DELETE FROM publications
WHERE publication_id NOT IN (SELECT pub_id FROM used_pubs);

-- Delete from reference_links
DELETE FROM reference_links
WHERE tsn NOT IN (SELECT tsn FROM selected_tsns);

-- Clean up unused lookup tables
DELETE FROM taxon_authors_lkp
WHERE taxon_author_id NOT IN (
  SELECT DISTINCT taxon_author_id FROM taxonomic_units
);

DELETE FROM taxon_unit_types
WHERE (rank_id, kingdom_id) NOT IN (
  SELECT DISTINCT rank_id, kingdom_id FROM taxonomic_units
);

-- Empty tables we don't use
DELETE FROM change_comments;
DELETE FROM change_operations;
DELETE FROM change_tracks;
DELETE FROM chg_operation_lkp;
DELETE FROM comments;
DELETE FROM experts;
DELETE FROM jurisdiction;
DELETE FROM longnames;
DELETE FROM nodc_ids;
DELETE FROM other_sources;
DELETE FROM reviews;
DELETE FROM strippedauthor;
DELETE FROM tu_comments_links;
DELETE FROM vern_ref_links;
DELETE FROM HierarchyToRank;

-- Vacuum to reclaim space
VACUUM;

-- Show final counts
SELECT 'taxonomic_units' as tbl, COUNT(*) as cnt FROM taxonomic_units
UNION ALL SELECT 'hierarchy', COUNT(*) FROM hierarchy
UNION ALL SELECT 'synonym_links', COUNT(*) FROM synonym_links
UNION ALL SELECT 'vernaculars', COUNT(*) FROM vernaculars
UNION ALL SELECT 'geographic_div', COUNT(*) FROM geographic_div
UNION ALL SELECT 'publications', COUNT(*) FROM publications
UNION ALL SELECT 'kingdoms', COUNT(*) FROM kingdoms
UNION ALL SELECT 'taxon_authors_lkp', COUNT(*) FROM taxon_authors_lkp
UNION ALL SELECT 'taxon_unit_types', COUNT(*) FROM taxon_unit_types;
