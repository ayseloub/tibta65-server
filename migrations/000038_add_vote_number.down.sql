DROP TRIGGER IF EXISTS trg_assign_vote_number ON votes;
DROP FUNCTION IF EXISTS assign_vote_number();
DROP INDEX IF EXISTS votes_vote_number_key;
ALTER TABLE votes DROP COLUMN IF EXISTS vote_number;