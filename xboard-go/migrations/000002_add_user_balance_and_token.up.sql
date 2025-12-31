-- Add balance, commission, and token fields to users table
-- These fields are for Xboard compatibility

ALTER TABLE users
    ADD COLUMN balance INT NOT NULL DEFAULT 0 COMMENT 'User balance in cents',
    ADD COLUMN discount INT NULL COMMENT 'User discount percentage',
    ADD COLUMN commission_type TINYINT NOT NULL DEFAULT 0 COMMENT '0: system 1: period 2: onetime',
    ADD COLUMN commission_rate INT NULL COMMENT 'Commission rate percentage',
    ADD COLUMN commission_balance INT NOT NULL DEFAULT 0 COMMENT 'Commission balance in cents',
    ADD COLUMN token VARCHAR(32) NULL COMMENT 'User API token',
    ADD COLUMN last_login_at TIMESTAMP NULL COMMENT 'Last login timestamp',
    ADD COLUMN last_login_ip VARCHAR(45) NULL COMMENT 'Last login IP address',
    ADD COLUMN remarks TEXT NULL COMMENT 'Admin remarks about user',
    ADD INDEX idx_token (token),
    ADD INDEX idx_last_login (last_login_at);
