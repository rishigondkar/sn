-- Create databases for each service.
-- Run with: psql -U cursor -h localhost -d postgres -f scripts/create_dbs.sql
-- (If a database already exists, that statement will error; others will still run.)

CREATE DATABASE case_db;
CREATE DATABASE alert_observable;
CREATE DATABASE enrichment_threat;
CREATE DATABASE assignment_ref;
CREATE DATABASE attachments;
CREATE DATABASE audit;
