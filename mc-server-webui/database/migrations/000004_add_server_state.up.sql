ALTER TABLE servers ADD COLUMN state TEXT DEFAULT 'online';
UPDATE servers SET state = 'offline' WHERE is_enabled = 0;
