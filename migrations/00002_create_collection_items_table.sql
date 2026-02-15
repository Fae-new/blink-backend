-- +goose Up
-- +goose StatementBegin
CREATE TABLE collection_items (
    id SERIAL PRIMARY KEY,
    collection_id INTEGER NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    parent_id INTEGER REFERENCES collection_items(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    item_type VARCHAR(20) NOT NULL CHECK (item_type IN ('folder', 'request')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    
    -- Request fields (NULL for folders)
    method VARCHAR(10) CHECK (method IN ('GET', 'POST', 'PUT', 'PATCH', 'DELETE')),
    url TEXT,
    headers JSONB,
    body TEXT,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_collection_items_collection_id ON collection_items(collection_id);
CREATE INDEX idx_collection_items_parent_id ON collection_items(parent_id);
CREATE INDEX idx_collection_items_sort_order ON collection_items(collection_id, parent_id, sort_order);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS collection_items;
-- +goose StatementEnd
