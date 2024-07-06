CREATE TABLE IF NOT EXISTS links (
	hash VARCHAR(6) PRIMARY KEY,
	original_url TEXT NOT NULL,
	correlation_id TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS correlation_id_idx ON links (correlation_id);