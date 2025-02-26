package itis

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gnames/coldp/ent/coldp"
)

func (i *itis) importNameUsage() error {
	batch := i.cfg.BatchSize
	q := `
SELECT h.TSN AS ID,
       h.Parent_TSN AS parentID,
	     tu.complete_name,
	     au.taxon_author,
       tu.rank_id,
       CASE WHEN tu.uncertain_prnt_ind = 'No'
	       THEN 'PROVISIONALLY_ACCEPTED' ELSE 'ACCEPTED' END AS status_id,
      (SELECT 
	       GROUP_CONCAT(expert, ',') FROM reference_links rl 
	         INNER JOIN experts 
	           ON rl.documentation_id=experts.expert_id 
	             AND rl.doc_id_prefix='EXP' WHERE rl.tsn=h.TSN) AS scrutinizer,
       tu.update_date AS scrutinizerDate,
       (SELECT
	        GROUP_CONCAT(documentation_id, ',')
	          FROM reference_links rl WHERE rl.tsn=h.TSN
	            AND doc_id_prefix='PUB') AS referenceID,
       0 AS extinct
FROM hierarchy h
LEFT JOIN taxonomic_units tu ON h.TSN = tu.tsn
LEFT JOIN taxon_authors_lkp au ON tu.taxon_author_id = au.taxon_author_id
	UNION
SELECT tu.TSN AS ID,
	   syn.tsn_accepted as parent_id,
	     tu.complete_name,
	     au.taxon_author,
       tu.rank_id,
       'SYNONYM' as status_id,
      (SELECT 
	       GROUP_CONCAT(expert, ',') FROM reference_links rl 
	         INNER JOIN experts 
	           ON rl.documentation_id=experts.expert_id 
	             AND rl.doc_id_prefix='EXP' WHERE rl.tsn=tu.TSN) AS scrutinizer,
       tu.update_date AS scrutinizerDate,
       (SELECT
	        GROUP_CONCAT(documentation_id, ',')
	          FROM reference_links rl WHERE rl.tsn=tu.TSN
	            AND doc_id_prefix='PUB') AS referenceID,
       0 AS extinct
FROM  taxonomic_units tu
LEFT JOIN taxon_authors_lkp au ON tu.taxon_author_id = au.taxon_author_id
INNER JOIN synonym_links syn on syn.tsn = tu.tsn
`
	rankMap, err := i.getRanks()
	if err != nil {
		return err
	}
	rows, err := i.itisDb.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()

	var status string
	var count int
	var rowsNum int64
	var scrut, ref, auth sql.NullString
	data := make([]coldp.NameUsage, 0, batch)
	for rows.Next() {
		count++
		rowsNum++
		var nu coldp.NameUsage
		var rank_id int
		err := rows.Scan(
			&nu.ID, &nu.ParentID, &nu.ScientificName, &auth, &rank_id, &status,
			&scrut, &nu.ScrutinizerDate, &ref, &nu.Extinct,
		)
		if err != nil {
			return err
		}
		nu.Authorship = auth.String
		nu.ScientificNameString = strings.TrimSpace(
			nu.ScientificName + " " + nu.Authorship,
		)
		nu.Scrutinizer = scrut.String
		nu.ReferenceID = ref.String
		nu.Rank = coldp.NewRank(rankMap[rank_id])
		nu.Link = "https://www.itis.gov/servlet/SingleRpt/" +
			"SingleRpt?search_topic=TSN&search_value=" + nu.ID

		data = append(data, nu)
		if count == batch {
			fmt.Printf("\r%s", strings.Repeat(" ", 50))
			fmt.Printf("\rProcessed %s lines", humanize.Comma(rowsNum))
			err := i.sfga.InsertNameUsages(data)
			if err != nil {
				return err
			}
			data = data[:0]
			count = 0
		}
	}
	fmt.Printf("\r%s\r", strings.Repeat(" ", 50))

	err = i.sfga.InsertNameUsages(data)
	if err != nil {
		return err
	}
	return nil
}

func (i *itis) getRanks() (map[int]string, error) {
	res := make(map[int]string)
	q := `SELECT DISTINCT rank_id, rank_name FROM taxon_unit_types`
	rows, err := i.itisDb.Query(q)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		}
		res[id] = name
	}
	return res, nil
}
