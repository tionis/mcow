CREATE TABLE IF NOT EXISTS servers (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name" TEXT NOT NULL UNIQUE,
    "address" TEXT NOT NULL,
    "description" TEXT,
    "blue_map_url" TEXT,
    "modpack_url" TEXT,
    "is_enabled" INTEGER DEFAULT 1
);
