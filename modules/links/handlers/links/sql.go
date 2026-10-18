package links

import "strings"

// insertLinkSQL - direct SQL insert for a link.
//
// Bun caches the struct schema and supplied an EMPTY value for link_code after
// the field was mapped to a renamed column, so the printed code was empty
// (/r/ with no code). Column/Value/ExcludeColumn/Safe did not help; keep this
// SQL in sync with the link_core migration.
//
// Placeholders are '?' (the pg driver in this stack is go-pg, NOT lib/pq:
// numbered $N placeholders fail there with "there is no parameter $1").
var insertLinkSQL = "INSERT INTO links (link_code, title, app, target_type, target_id, " +
	"external_url, path, enabled, redirect_kind, expires_at, note) VALUES (" +
	strings.TrimSuffix(strings.Repeat("?, ", 11), ", ") +
	") RETURNING id, hits, created_at, updated_at"
