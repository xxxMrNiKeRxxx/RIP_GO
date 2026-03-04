-- Добавление sort_order и is_primary в fuel_consumption_modes
ALTER TABLE fuel_consumption_modes
    ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0;

ALTER TABLE fuel_consumption_modes
    ADD COLUMN IF NOT EXISTS is_primary INTEGER REFERENCES driving_modes(mode_id);

-- Поля route_distance, fuel_saved уже NOT NULL