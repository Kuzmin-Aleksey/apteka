package sphinx

import (
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"server/config"
	"server/internal/domain/aggregate"
	"strings"
)

type Searcher struct {
	db    *sql.DB
	query string
}

func NewSearcher(cfg *config.SphinxConfig) (*Searcher, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("tcp(%s)/", cfg.Addr))
	if err != nil {
		return nil, err
	}

	return &Searcher{
		db:    db,
		query: "SELECT Code, WEIGHT() AS relevance FROM " + cfg.Index + " WHERE MATCH('%s') AND StoreID=%d LIMIT 1000 OPTION field_weights=(Name=3, Desctiption=2, Producer=1);",
	}, nil
}

func (s *Searcher) Search(ctx context.Context, storeId int, searchString string) ([]aggregate.SearchResult, error) {
	searchString = escapeSphinxQuery.Replace(strings.ToLower(searchString))

	results := make([]aggregate.SearchResult, 0)
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(s.query, searchString, storeId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return results, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var result aggregate.SearchResult
		if err := rows.Scan(&result.Code, &result.Relevance); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

var escapeSphinxQuery = strings.NewReplacer(
	`\`, `\\\`,
	`"`, `\\"`,
	`'`, `\\'`,
	`(`, `\\(`,
	`)`, `\\)`,
	`|`, `\\|`,
	`-`, `\\-`,
	`!`, `\\!`,
	`@`, `\\@`,
	`~`, `\\~`,
	`&`, `\\&`,
	`/`, `\\/`,
)
