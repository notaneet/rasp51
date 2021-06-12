package model

import "time"

// Class модель занятия
type Class struct {
	Title     string    `json:"title"`      //Название предмета (иногда с метадатой).
	StartTime time.Time `json:"start_time"` //Время начала пары
	EndTime   time.Time `json:"end_time"`   //Время конца пары
	Lecturer  string    `json:"lecturer"`   //Преподаватель
	Campus    string    `json:"campus"`     //Корпус с аудиторией/место
}
