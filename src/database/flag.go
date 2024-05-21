package database

import (
	"time"

	"github.com/hiumee/addack/src/model"
)

func (db *Database) CreateFlagsTable() error {
	_, err := db.DB.Exec("CREATE TABLE IF NOT EXISTS flags (id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, flag TEXT, exploit_id INTEGER, target_id INTEGER, result TEXT, valid TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		return err
	}
	_, err = db.DB.Exec("CREATE INDEX IF NOT EXISTS flag_index ON flags (flag)")
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

func (db *Database) FlagExists(flag string) bool {
	var id int64
	err := db.DB.QueryRow("SELECT id FROM flags WHERE flag = $1", flag).Scan(&id)
	if err != nil {
		return false
	}
	return true
}

func (db *Database) GetFlags(timezone string, timeformat string) ([]model.FlagDTO, error) {
	var flags []model.FlagDTO
	rows, err := db.DB.Query("SELECT flags.id, flag, exploits.name, targets.name, valid, timestamp FROM flags INNER JOIN exploits ON exploits.id = flags.exploit_id INNER JOIN targets ON targets.id = flags.target_id ORDER BY flags.id DESC LIMIT 100")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var flag model.FlagDTO
		var timestamp string
		err := rows.Scan(&flag.Id, &flag.Flag, &flag.ExploitName, &flag.TargetName, &flag.Valid, &timestamp)
		if err != nil {
			return nil, err
		}
		flagTime, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, err
		}
		flag.Timestamp = flagTime.In(location).Format(timeformat)

		flags = append(flags, flag)
	}

	return flags, nil
}

func (db *Database) GetMatchedFlags() []model.Flag {
	return db.GetFlagsCustomQuery("SELECT id, flag, exploit_id, target_id, result, valid FROM flags WHERE valid = 'matched'")
}

func (db *Database) SearchFlags(timezone string, timeformat string, exploit string, target string, flag string, valid string, content string) ([]model.FlagDTO, error) {
	var flags []model.FlagDTO

	exploit = "%" + exploit + "%"
	target = "%" + target + "%"
	flag = "%" + flag + "%"
	valid = valid + "%"
	content = "%" + content + "%"

	if valid == "" {
		valid = "%"
	}

	rows, err := db.DB.Query(`
		SELECT flags.id, flag, exploits.name, targets.name, valid, timestamp FROM flags 
			INNER JOIN exploits ON exploits.id = flags.exploit_id
			INNER JOIN targets ON targets.id = flags.target_id
		WHERE
			exploits.name LIKE $1 AND
			targets.name LIKE $2 AND
			flags.flag LIKE $3 AND
			valid LIKE $4 AND
			result LIKE $5
		ORDER BY flags.id DESC LIMIT 100
		`, exploit, target, flag, valid, content)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var flag model.FlagDTO
		var timestamp string
		err := rows.Scan(&flag.Id, &flag.Flag, &flag.ExploitName, &flag.TargetName, &flag.Valid, &timestamp)
		if err != nil {
			return nil, err
		}
		flagTime, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, err
		}
		flag.Timestamp = flagTime.In(location).Format(timeformat)

		flags = append(flags, flag)
	}

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

func (db *Database) UpdateFlagStatus(id int64, status string) error {
	_, err := db.DB.Exec("UPDATE flags SET valid = $1 WHERE id = $2", status, id)
	return err
}
