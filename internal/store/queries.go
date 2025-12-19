package store

// Credentials queries
const (
	queryGetCredentials = `
		SELECT url, username, password, is_data_sharing_allowed, created_at, updated_at
		FROM credentials WHERE id = 1`

	queryUpsertCredentials = `
		INSERT INTO credentials (id, url, username, password, is_data_sharing_allowed, updated_at)
		VALUES (1, ?, ?, ?, ?, now())
		ON CONFLICT (id) DO UPDATE SET
			url = EXCLUDED.url,
			username = EXCLUDED.username,
			password = EXCLUDED.password,
			is_data_sharing_allowed = EXCLUDED.is_data_sharing_allowed,
			updated_at = now()`

	queryDeleteCredentials = `DELETE FROM credentials WHERE id = 1`
)

// Inventory queries
const (
	queryGetInventory = `
		SELECT data, created_at, updated_at
		FROM inventory WHERE id = 1`

	queryUpsertInventory = `
		INSERT INTO inventory (id, data, updated_at)
		VALUES (1, ?, now())
		ON CONFLICT (id) DO UPDATE SET
			data = EXCLUDED.data,
			updated_at = now()`
)
