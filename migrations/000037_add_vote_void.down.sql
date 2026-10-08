ALTER TABLE votes DROP CONSTRAINT votes_member_id_fkey;
ALTER TABLE votes ADD CONSTRAINT votes_member_id_fkey
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE;
DROP TABLE IF EXISTS vote_void_logs;
ALTER TABLE votes DROP COLUMN IF EXISTS voided_at;