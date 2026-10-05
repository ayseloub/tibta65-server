ALTER TABLE members ADD COLUMN rejection_type VARCHAR(20) NULL;
ALTER TABLE members ADD COLUMN rejection_reason TEXT NULL;
ALTER TABLE members ADD COLUMN rejected_at TIMESTAMPTZ NULL;

UPDATE members SET status = 'incomplete' WHERE status = 'pending_review' AND profile_completed = false;