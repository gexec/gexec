package store

import (
	"context"

	"github.com/Machiel/slugify"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

// Slugify generates a slug.
func Slugify(_ context.Context, _ *bun.SelectQuery, column, value, id string, force bool) string {
	log.Debug().
		Str("column", column).
		Str("value", value).
		Str("id", id).
		Bool("force", force).
		Msg("Try to slug a value")

	// var (
	// 	slug string
	// )

	// for i := 0; true; i++ {
	// 	if i == 0 && !force {
	// 		slug = slugify.Slugify(value)
	// 	} else {
	// 		slug = slugify.Slugify(
	// 			fmt.Sprintf("%s-%s", value, uniuri.NewLen(6)),
	// 		)
	// 	}

	// 	query := db.ColumnExpr(
	// 		"? = ?",
	// 		bun.Ident(column),
	// 		slug,
	// 	)

	// 	if id != "" {
	// 		query = query.Where(
	// 			"id != ?",
	// 			id,
	// 		)
	// 	}

	// 	if count, err := query.Count(
	// 		ctx,
	// 	); err == nil && count == 0 {
	// 		break
	// 	}
	// }

	// return slug

	return slugify.Slugify(value)
}
