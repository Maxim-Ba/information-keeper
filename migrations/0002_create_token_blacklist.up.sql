CREATE TABLE token_blacklist (
    token TEXT PRIMARY KEY,
    expiry TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_token_blacklist_expiry ON token_blacklist(expiry);
CREATE INDEX idx_token_blacklist_token ON token_blacklist(token);
