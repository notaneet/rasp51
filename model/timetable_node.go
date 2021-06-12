package model

import "time"

//TimetableNode модель расписания
//Календарь с расписаниями вида день=>расписания
type TimetableNode map[time.Time][]Timetable
