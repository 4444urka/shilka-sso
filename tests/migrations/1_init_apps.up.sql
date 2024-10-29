INSERT INTO apps (id, name, secret)
VALUES (1, 'test', '4urka')
ON CONFLICT DO NOTHING;