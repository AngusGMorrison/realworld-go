CREATE OR REPLACE FUNCTION set_updated_at()
    RETURNS TRIGGER AS $$
        BEGIN
            NEW.updated_at = CURRENT_TIMESTAMP;
            RETURN NEW;
        END;
$$ language 'plpgsql';
