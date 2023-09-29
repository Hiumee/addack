package database

import (
	"github.com/hiumee/addack/src/model"
)

func (db *Database) CreateTargetsTable() error {
	_, err := db.DB.Exec("CREATE TABLE IF NOT EXISTS targets (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, name TEXT, ip TEXT, tag TEXT, enabled BOOLEAN DEFAULT 1);")
	return err
}

func (db *Database) CreateTarget(target model.Target) (int64, error) {
	res, err := db.DB.Exec("INSERT INTO targets (name, ip, tag, enabled) VALUES ($1, $2, $3, $4)", target.Name, target.Ip, target.Tag, target.Enabled)

	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()

	return id, nil
}

func (db *Database) GetTarget(id int64) (model.Target, error) {
	var target model.Target
	err := db.DB.QueryRow("SELECT * FROM targets WHERE id = $1", id).Scan(&target.Id, &target.Name, &target.Ip, &target.Tag, &target.Enabled)
	return target, err
}

func (db *Database) GetTargets() ([]model.Target, error) {
	var targets []model.Target
	rows, err := db.DB.Query("SELECT targets.id, targets.name, targets.ip, targets.tag, targets.enabled, count(flags.id) FROM targets LEFT JOIN flags ON targets.id = flags.target_id  AND flags.valid = 'valid' GROUP BY targets.id")
	if err != nil {
		return targets, err
	}
	defer rows.Close()
	for rows.Next() {
		var target model.Target
		err := rows.Scan(&target.Id, &target.Name, &target.Ip, &target.Tag, &target.Enabled, &target.Flags)
		if err != nil {
			return targets, err
		}
		targets = append(targets, target)
	}
	return targets, err
}

func (db *Database) UpdateTarget(target model.Target) error {
	_, err := db.DB.Exec("UPDATE targets SET name = $1, ip = $2, tag = $3, enabled = $4 WHERE id = $5", target.Name, target.Ip, target.Tag, target.Enabled, target.Id)
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

func (db *Database) SetEnabledTarget(target model.Target) error {
	_, err := db.DB.Exec("UPDATE targets SET enabled = $1 WHERE id = $2", target.Enabled, target.Id)
	return err
}
