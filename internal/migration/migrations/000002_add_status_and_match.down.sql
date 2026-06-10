-- SQLite does not support DROP COLUMN directly in older versions.
-- modernc.org/sqlite (v1.52+) supports it, so this is safe.
ALTER TABLE job_descriptions DROP COLUMN status;
ALTER TABLE job_descriptions DROP COLUMN fit_score;
ALTER TABLE job_descriptions DROP COLUMN fit_summary;
ALTER TABLE job_descriptions DROP COLUMN applied_at;
