-- Удаление поля total_saved из fuel_consumptions (если нужно)
ALTER TABLE fuel_consumptions DROP COLUMN IF EXISTS total_saved;