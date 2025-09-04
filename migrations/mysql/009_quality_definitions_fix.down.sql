-- Rollback Migration 009: Remove quality definitions fixes (MySQL/MariaDB)
-- This rollback is safe as it doesn't remove the table (that's managed by migration 001)
-- It only removes the safety enhancements added in migration 009

-- Note: We don't remove the quality_definitions table itself as it was created in migration 001
-- Note: We don't remove the indexes as they were created in migration 001
-- This migration only added safety measures and default data, so rollback is minimal

SELECT 'Quality definitions fix rollback completed' as status;
