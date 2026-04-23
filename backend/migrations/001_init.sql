-- ============================================
-- Bazzar Makuku - Event POS System
-- Migration 001: Initial Schema
-- ============================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- EVENTS (Multi-event support)
-- ============================================
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    start_date DATE,
    end_date DATE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- USERS
-- ============================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'picker')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- LOCATIONS (Event / Storage per event)
-- ============================================
CREATE TABLE locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id),
    code VARCHAR(50) NOT NULL, -- 'EVENT' or 'STORAGE'
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(event_id, code)
);

-- ============================================
-- SKUs (Master product data)
-- ============================================
CREATE TABLE skus (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sku_code VARCHAR(100) NOT NULL UNIQUE,  -- Shopee's "Nomor Referensi SKU"
    barcode VARCHAR(100) UNIQUE,             -- Physical barcode for scanning
    name VARCHAR(500) NOT NULL,
    description TEXT,
    replenish_limit INT DEFAULT 5,           -- Warning threshold
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- INVENTORY (Stock per SKU per Location)
-- ============================================
CREATE TABLE inventory (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sku_id UUID NOT NULL REFERENCES skus(id),
    location_id UUID NOT NULL REFERENCES locations(id),
    qty_onhand INT DEFAULT 0,      -- Physical stock at location
    qty_allocated INT DEFAULT 0,   -- Reserved for orders
    -- Available = qty_onhand - qty_allocated (computed)
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(sku_id, location_id)
);

-- ============================================
-- INBOUND PURCHASE ORDERS
-- ============================================
CREATE TABLE inbound_orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id),
    reference_number VARCHAR(100) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'partial', 'completed', 'cancelled')),
    notes TEXT,
    imported_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(event_id, reference_number)
);

CREATE TABLE inbound_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    inbound_order_id UUID NOT NULL REFERENCES inbound_orders(id) ON DELETE CASCADE,
    sku_id UUID NOT NULL REFERENCES skus(id),
    qty_expected INT NOT NULL DEFAULT 0,
    qty_received INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(inbound_order_id, sku_id)
);

-- ============================================
-- ORDERS (From Shopee import)
-- ============================================
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id),
    order_number VARCHAR(100) NOT NULL,         -- "No. Pesanan"
    platform_status VARCHAR(100),                -- Original Shopee status
    status VARCHAR(20) DEFAULT 'imported' CHECK (status IN (
        'imported', 'allocated', 'printed', 'picking', 'picked', 'shipped', 'cancelled', 'issue'
    )),
    buyer_name VARCHAR(255),                     -- "Nama Penerima"
    buyer_username VARCHAR(255),                 -- "Username (Pembeli)"
    shipping_option VARCHAR(255),                -- "Opsi Pengiriman"
    tracking_number VARCHAR(255),                -- "No. Resi"
    product_name VARCHAR(1000),                  -- "Nama Produk"
    variation_name VARCHAR(500),                 -- "Nama Variasi"
    notes TEXT,                                   -- "Catatan dari Pembeli"
    total_payment DECIMAL(15,2),                 -- "Total Pembayaran"
    assigned_picker_id UUID REFERENCES users(id),
    imported_by UUID REFERENCES users(id),
    printed_by UUID REFERENCES users(id),
    picked_by UUID REFERENCES users(id),
    shipped_by UUID REFERENCES users(id),
    imported_at TIMESTAMPTZ DEFAULT NOW(),
    allocated_at TIMESTAMPTZ,
    printed_at TIMESTAMPTZ,
    picking_started_at TIMESTAMPTZ,
    picked_at TIMESTAMPTZ,
    shipped_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(event_id, order_number)
);

-- ============================================
-- ORDER ITEMS
-- ============================================
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    sku_id UUID REFERENCES skus(id),              -- Linked SKU (nullable if unknown SKU)
    sku_code VARCHAR(100) NOT NULL,               -- Original "Nomor Referensi SKU"
    product_name VARCHAR(1000),
    variation_name VARCHAR(500),
    qty_ordered INT NOT NULL DEFAULT 1,
    qty_picked INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- INVENTORY LOGS (Full audit trail)
-- ============================================
CREATE TABLE inventory_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sku_id UUID NOT NULL REFERENCES skus(id),
    location_id UUID NOT NULL REFERENCES locations(id),
    event_id UUID NOT NULL REFERENCES events(id),
    action VARCHAR(30) NOT NULL CHECK (action IN (
        'inbound', 'allocate', 'deallocate', 'pick', 'ship', 
        'replenish_out', 'replenish_in', 'adjust', 'return'
    )),
    qty_change INT NOT NULL,                      -- Positive or negative
    reference_id UUID,                             -- order_id, inbound_id, etc.
    reference_type VARCHAR(30),                    -- 'order', 'inbound', 'replenish', 'manual'
    user_id UUID REFERENCES users(id),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- ACTIVITY LOG (All user actions)
-- ============================================
CREATE TABLE activity_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    event_id UUID REFERENCES events(id),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),                       -- 'order', 'inventory', 'inbound', etc.
    entity_id UUID,
    details JSONB,
    ip_address VARCHAR(45),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- INDEXES
-- ============================================
CREATE INDEX idx_orders_event_status ON orders(event_id, status);
CREATE INDEX idx_orders_order_number ON orders(order_number);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_sku_code ON order_items(sku_code);
CREATE INDEX idx_inventory_sku_location ON inventory(sku_id, location_id);
CREATE INDEX idx_inventory_logs_sku ON inventory_logs(sku_id);
CREATE INDEX idx_inventory_logs_created ON inventory_logs(created_at);
CREATE INDEX idx_activity_logs_user ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_created ON activity_logs(created_at);
CREATE INDEX idx_inbound_items_order ON inbound_items(inbound_order_id);
CREATE INDEX idx_skus_barcode ON skus(barcode);

-- ============================================
-- SEED DATA
-- ============================================
-- Default admin user (password: admin123)
INSERT INTO users (username, password_hash, full_name, role) VALUES
    ('admin', '$2a$10$placeholder_will_be_set_by_app', 'Administrator', 'admin');

-- Default event
INSERT INTO events (name, description, start_date, end_date) VALUES
    ('Bazzar Makuku', 'Bazzar Makuku Event', '2026-04-23', '2026-04-30');
