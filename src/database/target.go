package database

import (
	"addack/src/model"
)

func (db *Database) CreateTargetsTable() error {
	_, err := db.DB.Exec("CREATE TABLE IF NOT EXISTS targets (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, name TEXT, ip TEXT, tag TEXT);")
	return err
}

func (db *Database) CreateTarget(target model.Target) (int64, error) {
	res, err := db.DB.Exec("INSERT INTO targets (name, ip, tag) VALUES ($1, $2, $3)", target.Name, target.Ip, target.Tag)

	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()

	return id, nil
}

func (db *Database) GetTarget(id int64) (model.Target, error) {
	var target model.Target
	err := db.DB.QueryRow("SELECT * FROM targets WHERE id = $1", id).Scan(&target.Id, &target.Name, &target.Ip, &target.Tag)
	return target, err
}

func (db *Database) GetTargets() ([]model.Target, error) {
	var targets []model.Target
	rows, err := db.DB.Query("SELECT * FROM targets")
	if err != nil {
		return targets, err
	}
	defer rows.Close()
	for rows.Next() {
		var target model.Target
		err := rows.Scan(&target.Id, &target.Name, &target.Ip, &target.Tag)
		if err != nil {
			return targets, err
		}
		targets = append(targets, target)
	}
	return targets, err
}

func (db *Database) UpdateTarget(target model.Target) error {
	_, err := db.DB.Exec("UPDATE targets SET name = $1, ip = $2, tag = $3 WHERE id = $4", target.Name, target.Ip, target.Tag, target.Id)
	return err
}

func (db *Database) DeleteTarget(id int64) error {
	_, err := db.DB.Exec("DELETE FROM targets WHERE id = $1", id)
	return err
}

func (db *Database) DeleteAllTargets() error {
	_, err := db.DB.Exec("DELETE FROM targets")
	return err
}
