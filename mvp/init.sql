CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create the database if it doesn't exist
CREATE DATABASE smarthome;

-- Connect to the database
\c smarthome;

-- Таблица locations
CREATE TABLE locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL
);

-- Таблица sensor_types
CREATE TABLE sensor_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    code TEXT NOT NULL UNIQUE
);

-- Таблица sensors
CREATE TABLE sensors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    serial_number TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    last_activity_at TIMESTAMP WITHOUT TIME ZONE
);

-- Таблица sensor_heating
CREATE TABLE sensor_heating (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    serial_number TEXT NOT NULL UNIQUE REFERENCES sensors(serial_number) ON DELETE CASCADE,
    temperature NUMERIC NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

-- Таблица sensor_lighting
CREATE TABLE sensor_lighting (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    serial_number TEXT NOT NULL UNIQUE REFERENCES sensors(serial_number) ON DELETE CASCADE,
    status TEXT NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

-- Таблица sensor_gates
CREATE TABLE sensor_gates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    serial_number TEXT NOT NULL UNIQUE REFERENCES sensors(serial_number) ON DELETE CASCADE,
    status TEXT NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

-- Таблица telemetry
CREATE TABLE telemetry (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    serial_number TEXT NOT NULL,
    value JSONB NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

-- Индексы
CREATE INDEX idx_sensors_serial_number ON sensors(serial_number);
CREATE INDEX idx_sensor_heating_serial_number ON sensor_heating(serial_number);
CREATE INDEX idx_sensor_lighting_serial_number ON sensor_lighting(serial_number);
CREATE INDEX idx_sensor_gates_serial_number ON sensor_gates(serial_number);

-- Индексы для telemetry
CREATE INDEX idx_telemetry_serial_number ON telemetry(serial_number);
CREATE INDEX idx_telemetry_created_at ON telemetry(created_at);
CREATE INDEX idx_telemetry_serial_number_created_at ON telemetry(serial_number, created_at);