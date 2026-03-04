-- Удаляем таблицы в правильном порядке (от дочерних к родительским)
-- CASCADE запрещён по требованиям лабы

-- 1. Сначала таблица М-М связи (зависит от fuel_consumptions и driving_modes)
DROP TABLE IF EXISTS fuel_consumption_modes;

-- 2. Затем таблица заявок (зависит от users)
DROP TABLE IF EXISTS fuel_consumptions;

-- 3. Таблица режимов (независимая)
DROP TABLE IF EXISTS driving_modes;

-- 4. Последняя таблица пользователей (независимая)
DROP TABLE IF EXISTS users;