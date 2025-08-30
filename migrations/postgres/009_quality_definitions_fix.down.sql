-- Rollback Migration 009: Remove quality definitions fixes
-- This rollback is safe as it doesn't remove the table (that's managed by migration 001)
-- It only removes the safety enhancements added in migration 009

-- Drop the trigger created in this migration
DROP TRIGGER IF EXISTS update_quality_definitions_updated_at ON quality_definitions;
DROP FUNCTION IF EXISTS update_quality_definitions_updated_at();

-- Note: We don't remove the quality_definitions table itself as it was created in migration 001
-- Note: We don't remove the indexes as they were created in migration 001
-- This migration only added safety measures, so rollback is just removing the trigger