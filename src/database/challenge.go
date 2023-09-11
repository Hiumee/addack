package database

import (
	"addack/src/model"
)

// CreateChallenge creates a new challenge in the database
func (db *Database) CreateChallengesTable() error {
	_, err := db.DB.Exec("CREATE TABLE IF NOT EXISTS challenges (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, name TEXT, command TEXT, path TEXT, tag TEXT)")
	return err
}

// CreateChallenge creates a new challenge in the database
func (db *Database) CreateChallenge(challenge model.Challenge) (int64, error) {
	res, err := db.DB.Exec("INSERT INTO challenges (name, command, path, tag) VALUES ($1, $2, $3, $4)", challenge.Name, challenge.Command, challenge.Path, challenge.Tag)

	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()

	return id, nil
}

// GetChallenge returns a challenge from the database
func (db *Database) GetChallenge(id int64) (model.Challenge, error) {
	var challenge model.Challenge
	err := db.DB.QueryRow("SELECT * FROM challenges WHERE id = $1", id).Scan(&challenge.Id, &challenge.Name, &challenge.Command, &challenge.Path, &challenge.Tag)
	return challenge, err
}

// GetChallenges returns all challenges from the database
func (db *Database) GetChallenges() ([]model.Challenge, error) {
	var challenges []model.Challenge
	rows, err := db.DB.Query("SELECT * FROM challenges")
	if err != nil {
		return challenges, err
	}
	defer rows.Close()
	for rows.Next() {
		var challenge model.Challenge
		err := rows.Scan(&challenge.Id, &challenge.Name, &challenge.Command, &challenge.Path, &challenge.Tag)
		if err != nil {
			return challenges, err
		}
		challenges = append(challenges, challenge)
	}
	return challenges, err
}

// UpdateChallenge updates a challenge in the database
func (db *Database) UpdateChallenge(challenge model.Challenge) error {
	_, err := db.DB.Exec("UPDATE challenges SET name = $1, command = $2, path = $3, tag = $4 WHERE id = $5", challenge.Name, challenge.Command, challenge.Path, challenge.Tag, challenge.Id)
	return err
}

// DeleteChallenge deletes a challenge from the database
func (db *Database) DeleteChallenge(id int64) error {
	_, err := db.DB.Exec("DELETE FROM challenges WHERE id = $1", id)
	return err
}

func (db *Database) DeleteAllChallenges() error {
	_, err := db.DB.Exec("DELETE FROM challenges")
	return err
}
