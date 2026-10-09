ALTER TABLE votes ADD COLUMN vote_number INT NULL;

WITH numbered AS (
    SELECT id, ROW_NUMBER() OVER (ORDER BY created_at) AS rn FROM votes
)
UPDATE votes v SET vote_number = n.rn FROM numbered n WHERE v.id = n.id;

ALTER TABLE votes ALTER COLUMN vote_number SET NOT NULL;
CREATE UNIQUE INDEX votes_vote_number_key ON votes(vote_number);

CREATE FUNCTION assign_vote_number() RETURNS trigger AS $$
BEGIN
    PERFORM pg_advisory_xact_lock(7650001);
    SELECT COALESCE(MAX(vote_number), 0) + 1 INTO NEW.vote_number FROM votes;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_assign_vote_number
    BEFORE INSERT ON votes
    FOR EACH ROW EXECUTE FUNCTION assign_vote_number();