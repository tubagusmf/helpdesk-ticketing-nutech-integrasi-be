
-- +migrate Up
CREATE INDEX idx_ticket_project ON tickets(project_id);
CREATE INDEX idx_ticket_status ON tickets(status);
CREATE INDEX idx_ticket_priority ON tickets(priority);
CREATE INDEX idx_ticket_assigned ON tickets(assigned_to_id);
CREATE INDEX idx_ticket_reporter ON tickets(reporter_id);
CREATE INDEX idx_ticket_due_at ON tickets(due_at);

-- +migrate Down
DROP INDEX IF EXISTS idx_ticket_project;
DROP INDEX IF EXISTS idx_ticket_status;
DROP INDEX IF EXISTS idx_ticket_priority;
DROP INDEX IF EXISTS idx_ticket_assigned;
DROP INDEX IF EXISTS idx_ticket_reporter;
DROP INDEX IF EXISTS idx_ticket_due_at;