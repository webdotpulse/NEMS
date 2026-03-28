INSERT INTO devices (name, template, host, port, modbus_id, has_grid_meter, has_battery, battery_capacity) VALUES
('Demo Inverter', 'demo_inverter', 'localhost', 502, 1, 0, 1, 10.0),
('Demo Charger', 'demo_charger', 'localhost', 502, 1, 0, 0, 0.0),
('Demo Grid Meter', 'demo_dongle', 'localhost', 502, 1, 1, 0, 0.0);
