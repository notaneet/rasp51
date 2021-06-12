package model

import "time"

// Timetable модель одного учебного дня
type Timetable struct {
	Institution string    `json:"institution"` //Образовательное учреждение, к которой относится расписание
	GroupName   string    `json:"group"`       //Название группы
	Faculty     string    `json:"faculty"`     //Факультет
	Date        time.Time `json:"date"`        //День, к которому Timetable относится
	Activity    string    `json:"activity"`    //Какое-то задание дня (например, собрание, субботник, день самостоятельных работ, etc)
	Classes     []Class   `json:"classes"`     //Все занятия в этот день для этой группы
}
