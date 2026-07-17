CREATE TABLE IF NOT EXISTS task_labels (
    user_id  INTEGER NOT NULL,
    task_id  INTEGER NOT NULL,
    label_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, task_id, label_id),
    FOREIGN KEY (task_id, user_id) REFERENCES tasks (id, user_id) ON DELETE CASCADE,
    FOREIGN KEY (label_id, user_id) REFERENCES labels (id, user_id) ON DELETE CASCADE
);

CREATE INDEX idx_task_labels_task_id ON task_labels(task_id);
CREATE INDEX idx_task_labels_label_id ON task_labels(label_id);