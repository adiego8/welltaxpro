-- Portal Magic Link Tokens
-- Tracks one-time use magic link tokens for secure portal access

CREATE TABLE IF NOT EXISTS portal_magic_tokens (
    id UUID PRIMARY KEY,
    client_id UUID NOT NULL,
    tenant_id TEXT NOT NULL,
    email TEXT NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_magic_tokens_client ON portal_magic_tokens(client_id);
CREATE INDEX idx_magic_tokens_expiry ON portal_magic_tokens(expires_at);
CREATE INDEX idx_magic_tokens_used ON portal_magic_tokens(used);

-- Cleanup function to remove expired tokens (run daily)
CREATE OR REPLACE FUNCTION cleanup_expired_magic_tokens() RETURNS void AS $$
BEGIN
    DELETE FROM portal_magic_tokens
    WHERE expires_at < NOW() - INTERVAL '7 days';
END;
$$ LANGUAGE plpgsql;

COMMENT ON TABLE portal_magic_tokens IS 'Tracks one-time use magic link tokens with 24-hour expiry for portal access';
COMMENT ON COLUMN portal_magic_tokens.id IS 'JWT ID (jti claim) for token tracking';
COMMENT ON COLUMN portal_magic_tokens.used IS 'Marks if token has been exchanged for session token';
