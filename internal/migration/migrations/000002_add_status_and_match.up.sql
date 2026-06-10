ALTER TABLE job_descriptions ADD COLUMN status TEXT NOT NULL DEFAULT 'Draft';
ALTER TABLE job_descriptions ADD COLUMN fit_score INTEGER;
ALTER TABLE job_descriptions ADD COLUMN fit_summary TEXT;
ALTER TABLE job_descriptions ADD COLUMN applied_at DATETIME;
