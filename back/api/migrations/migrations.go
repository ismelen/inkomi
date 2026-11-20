package migrations

import _ "embed"

//go:embed users.sql
var Users string

//go:embed transactions.sql
var Transactions string

//go:embed joins.sql
var Joins string
