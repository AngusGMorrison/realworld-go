ALTER TABLE users
ADD COLUMN IF NOT EXISTS etag VARCHAR(255);

UPDATE users
SET etag = 'backfilled'
WHERE etag IS NULL;

ALTER TABLE users
ALTER COLUMN etag SET NOT NULL;