migrationUp:
	goose -dir db/migrations/postgres create first_migration sql