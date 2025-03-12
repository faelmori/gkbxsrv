-- docker/production/databases/init-db.sql

-- Extensões úteis
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- Para buscas de texto eficientes

-- Criação de roles e usuários
CREATE ROLE readonly;
CREATE ROLE readwrite;
CREATE ROLE admin;

-- Criação de usuários e atribuição de roles
CREATE USER user_readonly WITH PASSWORD 'readonlypass';
CREATE USER user_readwrite WITH PASSWORD 'readwritepass';
CREATE USER user_admin WITH PASSWORD 'adminpass';

GRANT readonly TO user_readonly;
GRANT readwrite TO user_readwrite;
GRANT admin TO user_admin;

-- Permissões para roles
GRANT CONNECT ON DATABASE shelfy TO readonly, readwrite, admin;
GRANT USAGE ON SCHEMA public TO readonly, readwrite, admin;

GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO readwrite;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO admin;

-- Enums
CREATE TYPE inventory_status AS ENUM ('available', 'reserved', 'damaged', 'expired');
CREATE TYPE order_status AS ENUM ('draft', 'pending', 'confirmed', 'processing', 'shipped', 'delivered', 'cancelled');
CREATE TYPE payment_status AS ENUM ('pending', 'paid', 'failed', 'refunded');
CREATE TYPE confidence_level AS ENUM ('high', 'medium', 'low');

-- Tabela de configurações de sincronização
CREATE TABLE sync_config (
                             id SERIAL PRIMARY KEY,
                             entity_name VARCHAR(100) NOT NULL,
                             last_sync_timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
                             sync_interval_minutes INTEGER NOT NULL DEFAULT 60,
                             is_active BOOLEAN NOT NULL DEFAULT TRUE,
                             error_count INTEGER NOT NULL DEFAULT 0,
                             created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                             updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Tabela de logs de sincronização
CREATE TABLE sync_logs (
                           id SERIAL PRIMARY KEY,
                           entity_name VARCHAR(100) NOT NULL,
                           start_time TIMESTAMP NOT NULL,
                           end_time TIMESTAMP,
                           status VARCHAR(50) NOT NULL,
                           records_processed INTEGER,
                           records_created INTEGER,
                           records_updated INTEGER,
                           records_failed INTEGER,
                           error_message TEXT,
                           created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Tabela de produtos
CREATE TABLE products (
                          id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                          external_id VARCHAR(100) UNIQUE, -- ID no Sankhya
                          sku VARCHAR(100) NOT NULL UNIQUE,
                          barcode VARCHAR(100),
                          name VARCHAR(255) NOT NULL,
                          description TEXT,
                          category VARCHAR(100),
                          manufacturer VARCHAR(100),
                          price DECIMAL(15, 2) DEFAULT 0,
                          cost DECIMAL(15, 2) DEFAULT 0,
                          weight DECIMAL(10, 3) DEFAULT 0,
                          length DECIMAL(10, 2) DEFAULT 0,
                          width DECIMAL(10, 2) DEFAULT 0,
                          height DECIMAL(10, 2) DEFAULT 0,
                          is_active BOOLEAN NOT NULL DEFAULT TRUE,
                          created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                          updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
                          last_sync_at TIMESTAMP NOT NULL DEFAULT NOW(),
    -- Campos de busca otimizados
                          search_vector tsvector,
    -- Campos específicos do Shelfy não presentes no Sankhya
                          min_stock_threshold INTEGER DEFAULT 0,
                          max_stock_threshold INTEGER DEFAULT 0,
                          reorder_point INTEGER DEFAULT 0,
                          lead_time_days INTEGER DEFAULT 0,
                          shelf_life_days INTEGER DEFAULT 0
);

-- Índices para produtos
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_barcode ON products(barcode);
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_manufacturer ON products(manufacturer);
CREATE INDEX idx_products_search_vector ON products USING GIN(search_vector);

-- Trigger para atualizar o campo search_vector
CREATE FUNCTION update_product_search_vector() RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
            setweight(to_tsvector('portuguese', COALESCE(NEW.name, '')), 'A') ||
            setweight(to_tsvector('portuguese', COALESCE(NEW.sku, '')), 'A') ||
            setweight(to_tsvector('portuguese', COALESCE(NEW.barcode, '')), 'A') ||
            setweight(to_tsvector('portuguese', COALESCE(NEW.description, '')), 'C') ||
            setweight(to_tsvector('portuguese', COALESCE(NEW.category, '')), 'B') ||
            setweight(to_tsvector('portuguese', COALESCE(NEW.manufacturer, '')), 'B');
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_product_search_vector
    BEFORE INSERT OR UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_product_search_vector();

-- Tabela de armazéns/localidades
CREATE TABLE warehouses (
                            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                            external_id VARCHAR(100) UNIQUE, -- ID no Sankhya
                            code VARCHAR(50) NOT NULL UNIQUE,
                            name VARCHAR(255) NOT NULL,
                            address TEXT,
                            city VARCHAR(100),
                            state VARCHAR(50),
                            country VARCHAR(50) DEFAULT 'Brasil',
                            postal_code VARCHAR(20),
                            is_active BOOLEAN NOT NULL DEFAULT TRUE,
                            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                            updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
                            last_sync_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Tabela de inventário
CREATE TABLE inventory (
                           id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                           product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
                           warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
                           quantity DECIMAL(15, 3) NOT NULL DEFAULT 0,
                           minimum_level DECIMAL(15, 3) DEFAULT 0,
                           maximum_level DECIMAL(15, 3) DEFAULT 0,
                           reorder_point DECIMAL(15, 3) DEFAULT 0,
                           last_count_date TIMESTAMP NOT NULL DEFAULT NOW(),
                           status inventory_status NOT NULL DEFAULT 'available',
                           lot_control VARCHAR(100) DEFAULT '0',
                           expiration_date TIMESTAMP,
                           location_code VARCHAR(100) DEFAULT '0',
                           created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                           updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
                           last_sync_at TIMESTAMP NOT NULL DEFAULT NOW(),
                           is_active BOOLEAN NOT NULL DEFAULT TRUE,
    -- Constraint para garantir combinação única de produto e armazém
                           CONSTRAINT unique_product_warehouse UNIQUE (product_id, warehouse_id, lot_control)
);

-- Índices para inventário
CREATE INDEX idx_inventory_product_id ON inventory(product_id);
CREATE INDEX idx_inventory_warehouse_id ON inventory(warehouse_id);
CREATE INDEX idx_inventory_status ON inventory(status);
CREATE INDEX idx_inventory_expiration_date ON inventory(expiration_date);

-- Tabela de movimentações de estoque
CREATE TABLE inventory_movements (
                                     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                     external_id VARCHAR(100) UNIQUE, -- ID no Sankhya
                                     inventory_id UUID NOT NULL REFERENCES inventory(id) ON DELETE CASCADE,
                                     product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
                                     warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
                                     quantity DECIMAL(15, 3) NOT NULL,
                                     previous_quantity DECIMAL(15, 3) NOT NULL,
                                     movement_type VARCHAR(50) NOT NULL, -- entrada, saída, ajuste, transferência
                                     reference_document VARCHAR(100), -- número do pedido, NF, etc.
                                     reason TEXT,
                                     created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                                     created_by VARCHAR(100),
                                     last_sync_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Índices para movimentações
CREATE INDEX idx_inventory_movements_inventory_id ON inventory_movements(inventory_id);
CREATE INDEX idx_inventory_movements_product_id ON inventory_movements(product_id);
CREATE INDEX idx_inventory_movements_warehouse_id ON inventory_movements(warehouse_id);
CREATE INDEX idx_inventory_movements_created_at ON inventory_movements(created_at);

-- Tabela de clientes (lojistas)
CREATE TABLE customers (
                           id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                           external_id VARCHAR(100) UNIQUE, -- ID no Sankhya
                           code VARCHAR(50) NOT NULL UNIQUE,
                           name VARCHAR(255) NOT NULL,
                           document VARCHAR(20) UNIQUE, -- CNPJ
                           email VARCHAR(255),
                           phone VARCHAR(20),
                           address TEXT,
                           city VARCHAR(100),
                           state VARCHAR(50),
                           country VARCHAR(50) DEFAULT 'Brasil',
                           postal_code VARCHAR(20),
                           is_active BOOLEAN NOT NULL DEFAULT TRUE,
                           created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                           updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
                           last_sync_at TIMESTAMP NOT NULL DEFAULT NOW(),
    -- Campos específicos do Shelfy
                           credit_limit DECIMAL(15, 2),
                           payment_terms VARCHAR(100),
                           last_purchase_date TIMESTAMP
);

-- Índices para clientes
CREATE INDEX idx_customers_code ON customers(code);
CREATE INDEX idx_customers_document ON customers(document);
CREATE INDEX idx_customers_name ON customers(name);
CREATE INDEX idx_customers_city_state ON customers(city, state);

-- Tabela de pedidos
CREATE TABLE orders (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        external_id VARCHAR(100) UNIQUE, -- ID no Sankhya
                        order_number VARCHAR(50) NOT NULL UNIQUE,
                        customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
                        status order_status NOT NULL DEFAULT 'draft',
                        order_date TIMESTAMP NOT NULL DEFAULT NOW(),
                        estimated_delivery_date TIMESTAMP,
                        actual_delivery_date TIMESTAMP,
                        shipping_address TEXT,
                        payment_method VARCHAR(100),
                        payment_status payment_status NOT NULL DEFAULT 'pending',
                        notes TEXT,
                        total_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
                        discount_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
                        tax_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
                        shipping_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
                        final_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
                        is_automatically_generated BOOLEAN NOT NULL DEFAULT FALSE,
                        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                        updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
                        last_sync_at TIMESTAMP NOT NULL DEFAULT NOW(),
    -- Campos específicos do Shelfy
                        prediction_id UUID, -- Referência à previsão que gerou este pedido (se automático)
                        priority INTEGER DEFAULT 0, -- Prioridade de processamento
                        expected_margin DECIMAL(15, 2) -- Margem esperada
);

-- Índices para pedidos
CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_order_date ON orders(order_date);
CREATE INDEX idx_orders_payment_status ON orders(payment_status);

-- Tabela de itens de pedido
CREATE TABLE order_items (
                             id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                             external_id VARCHAR(100) UNIQUE, -- ID no Sankhya
                             order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
                             product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
                             quantity DECIMAL(15, 3) NOT NULL,
                             unit_price DECIMAL(15, 2) NOT NULL,
                             discount DECIMAL(15, 2) NOT NULL DEFAULT 0,
                             total DECIMAL(15, 2) NOT NULL,
                             created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                             updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
                             last_sync_at TIMESTAMP NOT NULL DEFAULT NOW(),
    -- Campos específicos do Shelfy
                             is_suggested BOOLEAN DEFAULT FALSE, -- Se foi sugerido pelo sistema
                             suggestion_reason TEXT -- Razão da sugestão (ruptura iminente, promoção, etc.)
);

-- Índices para itens de pedido
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);

-- Tabela de previsões de estoque
CREATE TABLE stock_predictions (
                                   id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                   product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
                                   warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
                                   current_level DECIMAL(15, 3) NOT NULL,
                                   predicted_level DECIMAL(15, 3) NOT NULL,
                                   days_until_stockout INTEGER,
                                   confidence_level confidence_level NOT NULL,
                                   suggested_reorder_quantity DECIMAL(15, 3),
                                   prediction_date TIMESTAMP NOT NULL DEFAULT NOW(),
                                   prediction_horizon_days INTEGER NOT NULL DEFAULT 30,
                                   created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                                   updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    -- Constraint para garantir combinação única de produto, armazém e data
                                   CONSTRAINT unique_prediction UNIQUE (product_id, warehouse_id, prediction_date, prediction_horizon_days)
);

-- Índices para previsões
CREATE INDEX idx_stock_predictions_product_id ON stock_predictions(product_id);
CREATE INDEX idx_stock_predictions_warehouse_id ON stock_predictions(warehouse_id);
CREATE INDEX idx_stock_predictions_days_until_stockout ON stock_predictions(days_until_stockout);
CREATE INDEX idx_stock_predictions_confidence_level ON stock_predictions(confidence_level);

-- Tabela de dados de previsão diária (para armazenar séries temporais)
CREATE TABLE prediction_daily_data (
                                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                       prediction_id UUID NOT NULL REFERENCES stock_predictions(id) ON DELETE CASCADE,
                                       day_date DATE NOT NULL,
                                       predicted_demand DECIMAL(15, 3) NOT NULL,
                                       predicted_stock DECIMAL(15, 3) NOT NULL,
                                       lower_bound DECIMAL(15, 3), -- Limite inferior do intervalo de confiança
                                       upper_bound DECIMAL(15, 3), -- Limite superior do intervalo de confiança
                                       created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    -- Constraint para garantir combinação única de previsão e dia
                                       CONSTRAINT unique_prediction_day UNIQUE (prediction_id, day_date)
);

-- Índices para dados diários de previsão
CREATE INDEX idx_prediction_daily_data_prediction_id ON prediction_daily_data(prediction_id);
CREATE INDEX idx_prediction_daily_data_day_date ON prediction_daily_data(day_date);

-- Tabela de configurações de usuários
CREATE TABLE user_preferences (
                                  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                  user_id VARCHAR(100) NOT NULL,
                                  preference_key VARCHAR(100) NOT NULL,
                                  preference_value TEXT,
                                  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                                  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    -- Constraint para garantir combinação única de usuário e preferência
                                  CONSTRAINT unique_user_preference UNIQUE (user_id, preference_key)
);

-- Índices para preferências de usuários
CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);

-- Tabela de eventos de auditoria
CREATE TABLE audit_events (
                              id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                              entity_type VARCHAR(100) NOT NULL,
                              entity_id UUID NOT NULL,
                              action VARCHAR(50) NOT NULL, -- create, update, delete
                              user_id VARCHAR(100),
                              changes JSONB,
                              created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Índices para eventos de auditoria
CREATE INDEX idx_audit_events_entity_type_id ON audit_events(entity_type, entity_id);
CREATE INDEX idx_audit_events_created_at ON audit_events(created_at);
CREATE INDEX idx_audit_events_user_id ON audit_events(user_id);

insert into warehouses as w (external_id, code, name, address, city, state, country, postal_code, is_active)
values ('0','0','Warehouse','','','','','',true) on conflict do nothing;

-- Tabela temporária para inventário
CREATE TABLE temp_inventory (
                                id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                product_id UUID NOT NULL,
                                warehouse_id UUID,
                                quantity DECIMAL(15, 3) NOT NULL DEFAULT 0,
                                minimum_level DECIMAL(15, 3) DEFAULT 0,
                                maximum_level DECIMAL(15, 3) DEFAULT 0,
                                reorder_point DECIMAL(15, 3) DEFAULT 0,
                                last_count_date TIMESTAMP NOT NULL DEFAULT NOW(),
                                status inventory_status NOT NULL DEFAULT 'available',
                                lot_control VARCHAR(100) DEFAULT '0',
                                expiration_date TIMESTAMP,
                                location_code VARCHAR(100) DEFAULT '0',
                                created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                                updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
                                last_sync_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Trigger para validar e inserir dados na tabela final
CREATE FUNCTION validate_and_insert_inventory() RETURNS TRIGGER AS $$
BEGIN
    IF (NEW.warehouse_id IS NOT NULL AND EXISTS (SELECT 1 FROM warehouses WHERE id = NEW.warehouse_id AND is_active)) THEN
        INSERT INTO inventory (id, product_id, warehouse_id, quantity, minimum_level, maximum_level, reorder_point, last_count_date, status, lot_control, expiration_date, location_code, created_at, updated_at, last_sync_at)
        VALUES (NEW.id, NEW.product_id, NEW.warehouse_id, NEW.quantity, NEW.minimum_level, NEW.maximum_level, NEW.reorder_point, NEW.last_count_date, NEW.status, NEW.lot_control, NEW.expiration_date, NEW.location_code, NEW.created_at, NEW.updated_at, NEW.last_sync_at);
    ELSE
        INSERT INTO inventory (id, product_id, quantity, minimum_level, maximum_level, reorder_point, last_count_date, status, lot_control, expiration_date, location_code, created_at, updated_at, last_sync_at)
        VALUES (NEW.id, NEW.product_id, NEW.quantity, NEW.minimum_level, NEW.maximum_level, NEW.reorder_point, NEW.last_count_date, NEW.status, NEW.lot_control, NEW.expiration_date, NEW.location_code, NEW.created_at, NEW.updated_at, NEW.last_sync_at);
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_validate_and_insert_inventory
    AFTER INSERT ON temp_inventory
    FOR EACH ROW EXECUTE FUNCTION validate_and_insert_inventory();


COMMIT;
