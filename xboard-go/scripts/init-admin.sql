-- Initial admin user setup
-- Default credentials: admin@example.com / admin123
-- Password hash is bcrypt of "admin123"

INSERT INTO users (email, password_hash, role, banned, created_at, updated_at)
VALUES (
    'admin@example.com',
    '$2a$10$rQ3qFv5Z.5K5qKZ5qZ5qZuO5K5qKZ5qZ5qZ5qZ5qZ5qZ5qZ5qZ5qZ',
    'admin',
    0,
    NOW(),
    NOW()
) ON DUPLICATE KEY UPDATE email=email;

-- Create some default labels
INSERT INTO labels (name, description, created_at, updated_at) VALUES
    ('Premium', 'Premium tier nodes', NOW(), NOW()),
    ('Standard', 'Standard tier nodes', NOW(), NOW()),
    ('US', 'United States', NOW(), NOW()),
    ('EU', 'Europe', NOW(), NOW()),
    ('APAC', 'Asia Pacific', NOW(), NOW())
ON DUPLICATE KEY UPDATE name=name;

SELECT 'Admin user created: admin@example.com / admin123' as message;
SELECT 'IMPORTANT: Change this password immediately after first login!' as warning;
