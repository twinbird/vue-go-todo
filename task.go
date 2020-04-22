package main

import (
	"database/sql"
	"time"
)

const (
	StateAll = -1
	StateYet = 0
	StateDone
)

type Task struct {
	Id        int
	UserId    int
	Title     string
	State     int
	CreatedAt time.Time
}

func (t *Task) Create() error {
	var id int

	err := db.QueryRow(`
	INSERT INTO tasks(user_id, title, state, created_at)
	VALUES($1, $2, $3, $4)
	RETURNING id
	`, t.UserId, t.Title, StateYet, t.CreatedAt).Scan(&id)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	t.State = StateYet
	t.Id = id

	return nil
}

func (t *Task) Update() error {
	task, err := getTask(t.UserId, t.Id)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
	UPDATE
		tasks
	SET
		 title = $1
		,state = $2
	WHERE
		id = $3
	AND
		user_id = $4
	`, t.Title, t.State, task.Id, task.UserId)
	if err != nil {
		return err
	}
	task.Title = t.Title
	task.State = t.State
	return nil
}

func getTask(uid int, tid int) (*Task, error) {
	var task Task
	err := db.QueryRow(`
		SELECT
			 id
			,user_id
			,title
			,state
			,created_at
		FROM
			tasks
		WHERE
			user_id = $1
		AND
			id = $2
	`, uid, tid).Scan(&(task.Id), &(task.UserId), &(task.Title), &(task.State), &(task.CreatedAt))
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func getTasks(uid int, state int, q string) ([]Task, error) {
	query := `
		SELECT
			 id
			,user_id
			,title
			,state
			,created_at
		FROM
			tasks
		WHERE
			user_id = $1
		AND
			title LIKE $2
	`
	if state != StateAll {
		query += `
		AND
			state = $3`
	}
	query += `
		ORDER BY created_at ASC
	`

	var rows *sql.Rows
	var err error
	if state != StateAll {
		rows, err = db.Query(query, uid, "%"+q+"%", state)
	} else {
		rows, err = db.Query(query, uid, "%"+q+"%")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&(task.Id), &(task.UserId), &(task.Title), &(task.State), &(task.CreatedAt))
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
