CREATE TABLE todo_attatchments(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP(3) WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_At TIMESTAMP(3) WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    todo_id UUID NOT NULL REFERENCES todos ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    uploaded_by TEXT NOT NULL,
    download_key TEXT NOT NULL,
    file_size BIGINT,
    mime_type TEXT

);

CREATE INDEX idx_todo_attatchments_todo_id ON todo_attatchments(todo_id);
CREATE INDEX idx_todo_attatchments_uploaded_by ON todo_attatchments(uploaded_by);


CREATE TRIGGER set_todo_attatchments_update 
    BEFORE UPDATE on todo_attatchments
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

