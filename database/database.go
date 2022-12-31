package database

// package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// var DB *pgxpool.Pool
type DB struct {
	Pool *pgxpool.Pool
}

type User struct {
	Id         int64   `json:"user_id"`
	UserUuid   string  `json:"user_uuid"`
	UserName   string  `json:"user_name"`
	Ini        *string `json:"ini"`
	Mobile     *string `json:"mobile,omitempty"`
	College    *string `json:"college,omitempty"`
	Grade      *int64  `json:"grade,omitempty"`
	Email      string  `json:"email"`
	Registered bool    `json:"registered"`
}

type Team struct {
	Id       int64  `json:"-"`
	TeamName string `json:"team_name"`
	TeamCode string `json:"team_code"`
}

type Event struct {
	Id        int64  `json:"event_id"`
	EventName string `json:"event_name"`
	MinGrade  int64  `json:"min_grade"`
	MaxGrade  int64  `json:"max_grade"`
	MaxMates  int64  `json:"max_mates"`
}

type RegEventReq struct {
	EventId  int64   `json:"event_id"`
	TeamCode *string `json:"team_code"`
}

var DB_STRUCT *DB

func CreateDBPool(DATABASE_URL string) error {
	DB_STRUCT = new(DB)
	DB, err := pgxpool.New(context.Background(), DATABASE_URL)
	DB_STRUCT.Pool = DB

	if err != nil {
		return fmt.Errorf("pgx connection error: %w", err)
	}
	return nil
}

func (db *DB) GetUser(user_uuid string) (*User, error) {
	user := new(User)
	row := db.Pool.QueryRow(
		context.Background(),
		"SELECT id, user_uuid, user_name, mobile, college, grade, email, registered FROM registered_users WHERE user_uuid=$1",
		user_uuid,
	)

	err := row.Scan(
		&user.Id,
		&user.UserUuid,
		&user.UserName,
		&user.Mobile,
		&user.College,
		&user.Grade,
		&user.Email,
		&user.Registered,
	)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *DB) GetTeam(team_code *string) (*Team, error) {
	team := new(Team)
	row := db.Pool.QueryRow(
		context.Background(),
		"SELECT id, team_code, team_name FROM Team WHERE team_code=$1",
		team_code,
	)

	err := row.Scan(
		&team.Id,
		&team.TeamCode,
		&team.TeamName,
	)

	if err != nil {
		return nil, err
	}
	return team, nil
}

func (db *DB) GetEvent(event_id int64) (*Event, error) {
	event := new(Event)
	row := db.Pool.QueryRow(
		context.Background(),
		"SELECT id, event_name, min_grade, max_grade, max_mates FROM Event WHERE id=$1",
		event_id,
	)
	err := row.Scan(
		&event.Id,
		&event.EventName,
		&event.MinGrade,
		&event.MaxGrade,
		&event.MaxMates,
	)

	if err != nil {
		return nil, err
	}
	return event, nil
}

func (db *DB) CreateUser(user *User) (*User, error) {
	row := db.Pool.QueryRow(
		context.Background(),
		"INSERT INTO registered_users (user_uuid, user_name, email) VALUES ($1, $2, $3) RETURNING id",
		user.UserUuid, user.UserName, user.Email,
	)
	var user_id int64
	err := row.Scan(&user_id)
	if err != nil {
		return nil, err
	}
	user.Id = user_id
	return user, nil
}

func (db *DB) RegisterUser(user *User) (*User, error) {
	row := db.Pool.QueryRow(
		context.Background(),
		"UPDATE Registered_Users SET user_name=$1,college=$2,grade=$3,mobile=$4,registered=true WHERE id=$5 RETURNING 1",
		user.UserName, user.College, user.Grade, user.Mobile, user.Id,
	)
	var status int64
	err := row.Scan(&status)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *DB) CreateTeam(team *Team, initials *string) (*Team, error) {
	row := db.Pool.QueryRow(
		context.Background(),
		"INSERT INTO Team (team_name) VALUES ($1) RETURNING id",
		team.TeamName,
	)
	// var team_id int64
	err := row.Scan(&team.Id)
	if err != nil {
		return nil, fmt.Errorf("team name already exists")
	}
	team.TeamCode = fmt.Sprintf("CP%d%s", team.Id, *initials)
	row = db.Pool.QueryRow(
		context.Background(),
		"UPDATE TEAM SET team_code=$1 WHERE id=$2 RETURNING id",
		team.TeamCode, team.Id,
	)
	err = row.Scan(&team.Id)
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (db *DB) CheckUserEvent(event_id int64, user_id int64) error {
	var user_exists int64
	err := db.Pool.QueryRow(
		context.Background(),
		"SELECT 1 FROM EventTeam WHERE event_id=$1 AND user_id=$2",
		event_id, user_id,
	).Scan(&user_exists)

	if err != nil {
		return nil
	}
	return fmt.Errorf("user is already registered")
}

func (db *DB) GetTeamCount(event_id int64, team_id int64) int64 {
	var team_count int64
	err := db.Pool.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM EventTeam WHERE event_id=$1 AND team_id=$2",
		event_id, team_id,
	).Scan(&team_count)

	if err != nil {
		fmt.Println(err)
	}
	return team_count
}

func (db *DB) RegisterEventUser(req *RegEventReq, user *User) error {
	team, err := db.GetTeam(req.TeamCode)
	if err != nil {
		return fmt.Errorf("team code not found")
	}
	event, err := db.GetEvent(req.EventId)
	if err != nil {
		return fmt.Errorf("event not found")
	}
	if *user.Grade < event.MinGrade || *user.Grade > event.MaxGrade {
		return fmt.Errorf("age doesn't fit")
	}
	user_event_check := db.CheckUserEvent(req.EventId, user.Id)
	if user_event_check != nil {
		return user_event_check
	}

	team_count_check := db.GetTeamCount(req.EventId, team.Id)
	if team_count_check == event.MaxMates {
		return fmt.Errorf("team is full")
	}

	err = db.Pool.QueryRow(
		context.Background(),
		"INSERT INTO EventTeam (user_id,event_id,team_id) VALUES ($1,$2,$3) RETURNING 1",
		user.Id, event.Id, team.Id,
	).Scan(&team_count_check)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) RemoveEventUser(user_id int64, event_id int) error {
	_ = db.Pool.QueryRow(
		context.Background(),
		"DELETE FROM EventTeam WHERE user_id=$1 AND event_id=$2",
		user_id, event_id,
	)
	return nil
}
