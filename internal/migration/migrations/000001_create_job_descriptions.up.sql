CREATE TABLE job_descriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    company TEXT,
    role_title TEXT,
    seniority TEXT,
    employment_type TEXT,
    work_arrangement TEXT,
    location TEXT,
    requirements_json TEXT NOT NULL DEFAULT '{}',
    responsibilities_json TEXT NOT NULL DEFAULT '[]',
    keywords_json TEXT NOT NULL DEFAULT '[]',
    parsing_warning_json TEXT NOT NULL DEFAULT '[]',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
