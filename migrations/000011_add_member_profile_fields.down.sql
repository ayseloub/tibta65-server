ALTER TABLE members DROP CONSTRAINT IF EXISTS members_member_number_unique;
ALTER TABLE members DROP COLUMN IF EXISTS member_number;
ALTER TABLE members DROP COLUMN IF EXISTS address;
ALTER TABLE members DROP COLUMN IF EXISTS phone;
DROP SEQUENCE IF EXISTS member_number_seq;