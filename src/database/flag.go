package database

import (
	"addack/src/model"
	"log"
)

func (db *Database) CreateFlagsTable() error {
	_, err := db.DB.Exec("CREATE TABLE IF NOT EXISTS flags (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, flag TEXT, exploit_id INTEGER, target_id INTEGER, result TEXT, valid BOOLEAN)")
	return err
}

func (db *Database) CreateFlag(flag model.Flag) (int64, error) {
	res, err := db.DB.Exec("INSERT INTO flags (flag, exploit_id, target_id, result, valid) VALUES ($1, $2, $3, $4, $5)", flag.Flag, flag.ExploitId, flag.TargetId, flag.Result, flag.Valid)

	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()

	return id, nil
}

func (db *Database) GetFlags() ([]model.FlagDTO, error) {
	var flags []model.FlagDTO
	rows, err := db.DB.Query("SELECT flags.id, flag, exploits.name, targets.name, valid FROM flags INNER JOIN exploits ON exploits.id = flags.exploit_id INNER JOIN targets ON targets.id = flags.target_id ORDER BY flags.id DESC LIMIT 100")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var flag model.FlagDTO
		err := rows.Scan(&flag.Id, &flag.Flag, &flag.ExploitName, &flag.TargetName, &flag.Valid)
		if err != nil {
			return nil, err
		}
		flags = append(flags, flag)
	}
	log.Default().Println(flags)
	return flags, nil
}

func (db *Database) GetFlagsCustomQuery(query string) []model.Flag {
	var flags []model.Flag
	rows, err := db.DB.Query(query)
	if err != nil {
		return flags
	}
	defer rows.Close()
	for rows.Next() {
		var flag model.Flag
		err := rows.Scan(&flag.Id, &flag.Flag, &flag.ExploitId, &flag.TargetId, &flag.Result, &flag.Valid)
		if err != nil {
			return flags
		}
		flags = append(flags, flag)
	}
	return flags
}

func (db *Database) GetFlagResult(id int64) (string, error) {
	var result string
	err := db.DB.QueryRow("SELECT result FROM flags WHERE id = $1", id).Scan(&result)
	return result, err
}
