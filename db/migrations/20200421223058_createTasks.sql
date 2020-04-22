
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE tasks (
	id SERIAL PRIMARY KEY,
	user_id INTEGER REFERENCES users(id),
	title VARCHAR(100) NOT NULL,
	state INTEGER NOT NULL,
	created_at TIMESTAMP NOT NULL
);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE tasks;
