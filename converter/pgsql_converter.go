package converter

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/notaneet/rasp51/model"
)

type PGSQLConverter struct{}

const ResetSequence = "ALTER SEQUENCE timetables_id_seq RESTART;"
const DropClasses = "DELETE FROM classes;"
const DropTimetables = "DELETE FROM timetables;"
const InsertTimetableQuery = "INSERT INTO timetables (institution, date, \"group\", faculty, activity) VALUES ($1, $2, $3, $4, $5) RETURNING id"
const InsertClassQuery = "INSERT INTO classes (timetable_id, title, start_time, end_time, lecturer, campus) VALUES ($1, $2, $3, $4, $5, $6)"

func (p PGSQLConverter) Write(node model.TimetableNode, out string) error {
	if out == "" {
		return fmt.Errorf("credentials can not be empty")
	}

	conn, err := sqlx.Connect("postgres", out)
	if err != nil {
		return err
	}
	conn.MustExec(ResetSequence)

	conn.MustExec(DropClasses)
	conn.MustExec(DropTimetables)

	insertTimetable, err := conn.Preparex(InsertTimetableQuery)
	if err != nil {
		return err
	}

	insertClass, err := conn.Preparex(InsertClassQuery)
	if err != nil {
		return err
	}

	for _, days := range node {
		for _, day := range days {
			var timetableId uint
			scan := insertTimetable.QueryRowx(day.Institution, day.Date, day.GroupName, day.Faculty, day.Activity)
			if err = scan.Scan(&timetableId); err != nil {
				continue
			}

			for _, class := range day.Classes {
				insertClass.MustExec(timetableId, class.Title, class.StartTime, class.EndTime, class.Lecturer, class.Campus)
			}
		}
	}

	return nil
}