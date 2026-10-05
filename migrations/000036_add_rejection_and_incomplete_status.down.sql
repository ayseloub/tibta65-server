UPDATE members SET status = 'pending_review' WHERE status = 'incomplete';
ALTER TABLE members DROP COLUMN IF EXISTS rejected_at;
ALTER TABLE members DROP COLUMN IF EXISTS rejection_reason;
ALTER TABLE members DROP COLUMN IF EXISTS rejection_type;