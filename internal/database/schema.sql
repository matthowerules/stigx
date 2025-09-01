-- SQLite schema for STIG Viewer X
-- Database for storing imported STIG metadata and user data

-- STIGs table - stores metadata about imported STIGs
CREATE TABLE IF NOT EXISTS stigs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,                    -- e.g., "Apache Server 2.4 Unix Server"
    version TEXT NOT NULL,                 -- e.g., "V3R2"
    release_date TEXT,                     -- e.g., "2024-12-04"
    benchmark_id TEXT UNIQUE NOT NULL,     -- XCCDF benchmark ID
    title TEXT,                            -- Full title from XCCDF
    description TEXT,                      -- Description from XCCDF
    filename TEXT NOT NULL,                -- Original zip filename
    file_path TEXT,                        -- Path to extracted XCCDF
    file_hash TEXT,                        -- SHA-256 hash of XCCDF file
    rule_count INTEGER DEFAULT 0,          -- Number of rules in this STIG
    imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- STIG Rules table - stores individual rules from STIGs  
CREATE TABLE IF NOT EXISTS stig_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    stig_id INTEGER NOT NULL,
    rule_id TEXT NOT NULL,                 -- XCCDF rule ID
    version_id TEXT,                       -- Version identifier  
    title TEXT,                            -- Rule title
    description TEXT,                      -- Rule description
    severity TEXT,                         -- high, medium, low
    weight REAL,                           -- Weight/score
    group_id TEXT,                         -- Group this rule belongs to
    group_title TEXT,                      -- Group title
    check_content TEXT,                    -- How to check compliance
    fix_text TEXT,                         -- How to fix non-compliance
    cci_refs TEXT,                         -- CCI references (JSON array)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (stig_id) REFERENCES stigs(id) ON DELETE CASCADE,
    UNIQUE(stig_id, rule_id)
);

-- Bundles table - tracks imported bundle files
CREATE TABLE IF NOT EXISTS bundles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    filename TEXT NOT NULL,               -- e.g., "U_SRG-STIG_Library_April_2025.zip"
    file_path TEXT NOT NULL,              -- Path to bundle file
    file_hash TEXT NOT NULL,              -- SHA-256 hash of bundle
    stig_count INTEGER DEFAULT 0,         -- Number of STIGs extracted
    imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);



-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_stigs_benchmark_id ON stigs(benchmark_id);
CREATE INDEX IF NOT EXISTS idx_stigs_name_version ON stigs(name, version);
CREATE INDEX IF NOT EXISTS idx_stig_rules_stig_id ON stig_rules(stig_id);
CREATE INDEX IF NOT EXISTS idx_stig_rules_rule_id ON stig_rules(rule_id);

-- Views for convenience
CREATE VIEW IF NOT EXISTS stig_summary AS
SELECT
    s.id,
    s.name,
    s.version,
    s.release_date,
    s.title,
    s.rule_count,
    s.imported_at
FROM stigs s;