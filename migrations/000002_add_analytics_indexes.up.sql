CREATE INDEX idx_tasks_team_status_created ON tasks(team_id, status, created_at);

CREATE INDEX idx_tasks_team_creator_created ON tasks(team_id, created_by, created_at);
