
-- +migrate Up
CREATE TABLE app_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    app_name VARCHAR(100) NOT NULL DEFAULT 'Helpdesk Center',
    login_title VARCHAR(100) NOT NULL DEFAULT 'Helpdesk Center',
    tagline VARCHAR(200) NOT NULL DEFAULT 'Professional Ticketing System',
    logo_url TEXT NOT NULL DEFAULT '',
    login_icon_url TEXT NOT NULL DEFAULT '',
    favicon_url TEXT NOT NULL DEFAULT '',
    whatsapp_number VARCHAR(25) NOT NULL DEFAULT '6281119521124',
    whatsapp_display_number VARCHAR(40) NOT NULL DEFAULT '081119521124',
    whatsapp_message TEXT NOT NULL DEFAULT E'Halo Admin,\n\nMohon untuk mendaftar akun Helpdesk dengan mengisi data berikut:\n\nNama Lengkap:\nEmail:\nProject:\n\nTerima kasih.',
    primary_color VARCHAR(7) NOT NULL DEFAULT '#EA580C',
    sidebar_title VARCHAR(100) NOT NULL DEFAULT 'Helpdesk Center',
    footer_text TEXT NOT NULL DEFAULT '© 2026 Helpdesk CCIT Nutech Integrasi',
    app_version VARCHAR(30) NOT NULL DEFAULT 'Version 1.0.0',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO app_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

-- +migrate Down
DROP TABLE IF EXISTS app_settings;