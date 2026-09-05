CREATE SEQUENCE IF NOT EXISTS member_number_seq START 1;

ALTER TABLE members ADD COLUMN phone VARCHAR(20) NULL;
ALTER TABLE members ADD COLUMN address TEXT NULL;
ALTER TABLE members ADD COLUMN member_number VARCHAR(30) NULL;

UPDATE members
SET member_number = 'TIBTA-1965-' || LPAD(nextval('member_number_seq')::text, 4, '0')
WHERE member_number IS NULL;

ALTER TABLE members ALTER COLUMN member_number SET NOT NULL;
ALTER TABLE members ADD CONSTRAINT members_member_number_unique UNIQUE (member_number);