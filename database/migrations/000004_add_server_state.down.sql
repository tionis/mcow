UPDATE servers SET is_enabled = 0 WHERE state = 'offline';
ALTER TABLE servers DROP COLUMN state;
