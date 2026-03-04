CREATE TABLE IF NOT EXISTS users (
                                     user_id SERIAL PRIMARY KEY,
                                     login VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL,
    is_moderator BOOLEAN DEFAULT FALSE
    );

CREATE TABLE IF NOT EXISTS driving_modes (
                                             mode_id SERIAL PRIMARY KEY,
                                             mode_name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    image_key VARCHAR(255),
    video_key VARCHAR(255),
    base_consumption NUMERIC(5,2) NOT NULL,
    economy_percent NUMERIC(5,2) NOT NULL,
    driving_type VARCHAR(20) NOT NULL
    );

CREATE TABLE IF NOT EXISTS fuel_consumptions (
                                                 consumption_id SERIAL PRIMARY KEY,
                                                 application_id VARCHAR(20) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    creator_id INTEGER NOT NULL REFERENCES users(user_id),
    origin VARCHAR(100) NOT NULL,
    destination VARCHAR(100) NOT NULL,
    forming_date TIMESTAMP,
    finish_date TIMESTAMP,
    moderator_id INTEGER REFERENCES users(user_id),
    fuel_price NUMERIC(10,2) DEFAULT 55.00,
    total_saved NUMERIC(10,2)
    );

CREATE TABLE IF NOT EXISTS fuel_consumption_modes (
                                                      consumption_id INTEGER NOT NULL REFERENCES fuel_consumptions(consumption_id),
    mode_id INTEGER NOT NULL REFERENCES driving_modes(mode_id),
    route_distance INTEGER NOT NULL,
    fuel_saved NUMERIC(10,2) NOT NULL,
    is_primary INTEGER REFERENCES driving_modes(mode_id),
    sort_order INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (consumption_id, mode_id)
    );

-- Начальные данные
INSERT INTO users (login, password, is_moderator) VALUES
                                                      ('user1', 'pass1', false),
                                                      ('moderator', 'modpass', true)
    ON CONFLICT (login) DO NOTHING;

INSERT INTO driving_modes (mode_name, description, is_active, image_key, video_key, base_consumption, economy_percent, driving_type) VALUES
                                                                                                                                         ('Городской режим - Компактный', 'Расчет экономии топлива для городского режима движения.', TRUE, 'car_city_compact.jpg', 'car_city_compact.mp4', 8.00, 5.00, 'city'),
                                                                                                                                         ('Городской режим - Седан', 'Расчет экономии топлива для городского режима на седанах.', TRUE, 'car_city_sedan.jpg', 'car_city_sedan.mp4', 10.00, 5.00, 'city'),
                                                                                                                                         ('Трасса - Компактный', 'Расчет экономии топлива для трассы на компактных автомобилях.', TRUE, 'car_highway_compact.jpg', 'car_highway_compact.mp4', 6.00, 15.00, 'highway'),
                                                                                                                                         ('Трасса - Внедорожник', 'Расчет экономии топлива для трассы на внедорожниках.', TRUE, 'car_highway_suv.jpg', 'car_highway_suv.mp4', 12.00, 15.00, 'highway'),
                                                                                                                                         ('Смешанный режим - Седан', 'Расчет экономии топлива для смешанного режима движения.', TRUE, 'car_mixed_sedan.jpg', 'car_mixed_sedan.mp4', 9.00, 10.00, 'mixed'),
                                                                                                                                         ('Смешанный режим - Грузовой', 'Расчет экономии топлива для смешанного режима на грузовых автомобилях.', TRUE, 'car_mixed_truck.jpg', 'car_mixed_truck.mp4', 15.00, 10.00, 'mixed');