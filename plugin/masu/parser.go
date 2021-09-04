package masu

// НЕНАВИСТЬ!!!
// СПУСТЯ ГОД Я ВСЕ ПЕРЕСМОТРЕЛ И ПОНЯЛ СВОЮ ОШИБКУ
// НАХУЙ Я СТАЛ ДЕЛАТЬ УМНЫЙ ПАРСИНГ
// НЕЗАМОРАЧИВАЯСЬ БЫ ОТДАВАЛ ПРОСТО ВСЕ НЕ ПУСТЫЕ СТРОЧКИ
// ЕБАЛ В РОТ ЭТУ ХУЙНЮ
// за ма*т извените

//UPD: Это не я виноват, виноват отдел, составляющий расписание T_T

import (
	"fmt"
	"github.com/notaneet/rasp51/model"
	"github.com/notaneet/rasp51/utils"
	"github.com/tealeg/xlsx/v3"
	"math"
	"regexp"
	"strings"
	"time"
)

// Начинается всё с 10 строки
const startRow = 10

// И со 2 столбца
const startColumn = 2

// Кол-во пар максимальное - 7
const maxClasses = 7

// Длина одной пары в строках - 2
const classLength = 2

// Длина одного дня в столбцах - 2
const dayWidth = 2

// Если что-то не указано
const emptyField = "Не указан"

var crapyActivityHackRE = regexp.MustCompile("^с \\d{2}:\\d{2}$")
var crapyActivityHack2RE = regexp.MustCompile("^\\*+$")

// Спарсить xlsx файл
func (p *_MASUMurmanskPlugin) parseWB(wb *xlsx.File, faculty string) error {
	for _, sh := range wb.Sheets {
		name := cleanName(sh.Name)
		if len(name) == 0 {
			continue
		}

		group := currentGroup{
			groupNames: name,
			faculty:    faculty,
			sh:         sh,
		}

		err, days := p.parseSH(group)
		if err != nil {
			return err
		}

		for day, timetables := range days {
			for _, ret := range timetables {
				p.timetable[day] = append(p.timetable[day], ret)
			}
		}
	}
	return nil
}

// Спарсить лист в xlsx файле (учебную группу)
func (p *_MASUMurmanskPlugin) parseSH(group currentGroup) (error error, days map[time.Time][]model.Timetable) {
	//groupName не проходит
	for _, groupName := range group.groupNames {
		if !p.config.GroupMatcher.Match(groupName) {
			return nil, days
		}
	}

	for column := startColumn - 1; column < group.sh.MaxCol; column += dayWidth {
		for row := startRow - 1; row < group.sh.MaxRow; row += maxClasses*classLength + 1 {
			// Получим клетку, где по идее должна быть дата и если там нет клетки, либо там не дата, то идём ко след. дню
			dateCell, err := group.sh.Cell(row, column+1)
			if err != nil || !dateCell.IsTime() {
				continue
			}

			// Получим дату, и если ее нет, то переходим к след. дню
			date, err := dateCell.GetTime(false)
			if err != nil {
				continue
			}

			// Спарсим день
			err, t := p.parseDay(group, column, row, date)
			if err != nil {
				return err, nil
			}

			if days == nil {
				days = map[time.Time][]model.Timetable{}
			}

			for _, timetable := range t {
				if timetable != nil {
					days[date] = append(days[date], *timetable)
				}
			}
		}
	}
	return nil, days
}

//Почему в стд нет сплита с оффсетом, либо хотя-бы регулярок последней версии
func splitExcept(str string, separator, excepting *regexp.Regexp) (ret []string) {
	cleaned := excepting.ReplaceAllString(str, "<ugly_hack>")
	for _, s := range separator.Split(cleaned, -1) {
		ret = append(ret, strings.Replace(s, "<ugly_hack>", excepting.FindString(str), -1))
	}
	return
}

//Иногда у разных подгрупп разные пары в одно и тоже время (разделяется обычно с помощью //)
var subgroupSepartor = regexp.MustCompile("(?i)//")

//^ выше
func splitOnSubgroups(str string) []string {
	return splitExcept(str, subgroupSepartor, externalLinkRE)
}

//activity, где время не нужно
var timeIsUselessRE = regexp.MustCompile("(?i)(день самостоятельной|день самоподготовки|праздничный день|преддипломная практика|выходной|классный час|\\d{1,2}[.:]\\d{2})")

//Есть ли время в строке
var timeContainerRE = regexp.MustCompile("(?i)(\\d{1,2}[.:]\\d{2})")

// Спарсить какой-то день
// column - первый слева столбик (там, где день недели)
// row - первая с верху строчка (там, где день недели и дата)
func (p *_MASUMurmanskPlugin) parseDay(group currentGroup, column, row int, date time.Time) (error error, ret []*model.Timetable) {
	// Если передан интервал и дата за ее пределами, то ничего не возвращаем
	if (p.config.StartTime != nil && date.Before(*p.config.StartTime)) || (p.config.EndTime != nil && date.After(*p.config.EndTime)) {
		return nil, nil
	}
	// Костыль для колледжа...
	groups := group.groupNames

	// Занятия
	classes := make([]model.Class, 0)
	// Цель дня
	activity := ""
	// Возможное время начала
	possibleActivityStartTime := ""

	// Костыль, для пар, которые расположились на две строчки и имеют специфическое время.
	prevStartTime := (*time.Time)(nil)

	// Переберём все занятия
	for y := 1; y <= maxClasses; y++ {
		// Получим время занятия, если ошибка, вернем её
		timeCell, err := group.sh.Cell(row+y*2-1, column)
		if err != nil {
			return err, nil
		}

		// Разобъем строку по -
		// TODO: Может закостылить время начала и конца по номеру пары?
		rawSplited := strings.Split(timeCell.String(), "-")
		if len(rawSplited) < 2 {
			//return fmt.Errorf(timeCell.String() + " is not correct time string"), nil
			return nil, nil
		}

		// Получим время начала занятия
		startTime := utils.AddTimeToDate(date, rawSplited[0])
		// Если начало пары указано в прошлой паре, установим его
		if prevStartTime != nil {
			startTime = *prevStartTime
			prevStartTime = nil
		}

		// Получим время конца занятия
		endTime := utils.AddTimeToDate(date, rawSplited[1])

		// Получим первую строку (там обычно название пары), если там ошибка, то просто вернем её
		firstLineCell, err := group.sh.Cell(row+y*2-1, column+1)
		if err != nil {
			return err, nil
		}
		firstLine := firstLineCell.String()

		// Получим вторую строку (там обычно преподаватель и корпус), если там ошибка, просто вернем её
		secondLineCell, err := group.sh.Cell(row+y*2, column+1)
		if err != nil {
			return err, nil
		}
		secondLine := secondLineCell.String()

		// Если обе строки пустые, то идем дальше
		if len(firstLine) == 0 && len(secondLine) == 0 {
			continue
		}

		// Разобъем строки по под.группам и пройдемся по ним
		firstLineSplited := splitOnSubgroups(firstLine)
		secondLineSplited := splitOnSubgroups(secondLine)
		for line := 0; line < len(firstLineSplited); line++ {
			// Незаконченное занятие (которое сейчас и парсится)
			var lastClassEntry *model.Class = nil

			subgroups := resolveSubgroups(firstLineSplited, secondLineSplited, line)
			// Пройдёмся по всем под-группам
			for _, lines := range subgroups {
				firstSubLine, secondSubLine := lines[0], lines[1]
				if !p.config.ClassMatcher.Match(firstSubLine) {
					continue
				}

				// Если это занятие, то
				if isLecture(firstSubLine, secondLine) && len(secondSubLine) > 0 {
					// Разобъем вторую строку по /
					spliced := splitLine(secondSubLine)

					// Если указан и преподаватель и корпус, либо
					// Если кол-во данных во второй строчке равно 1 и нет пары до, то
					if len(spliced) == 2 || (len(spliced) == 1 && lastClassEntry == nil) {
						lecturer := lecturerName(strings.TrimSpace(strings.Replace(spliced[0], ",", "", -1)))

						campus := emptyField
						// Если указан корпус
						if len(spliced) == 2 {
							campus = campusName(strings.Replace(spliced[1], ",", " ", -1))
						} else if lecturer == emptyField { // Либо, если не указан препод
							campus = campusName(strings.TrimSpace(strings.Replace(spliced[0], ",", "", -1)))
						}

						if !p.config.LecturerMatcher.Match(lecturer) || !p.config.CampusMatcher.Match(campus) {
							continue
						}

						//TODO: Факинг колледж
						if campus == "*" || firstSubLine == "*" || firstSubLine == "" {
							continue
						}

						// Запишем текущее занятие, возможно будем его дополнять
						lastClassEntry = &model.Class{
							Title:     firstSubLine,
							StartTime: startTime,
							EndTime:   endTime,
							Lecturer:  lecturer,
							Campus:    campus,
						}
						// Время начала теперь не нужно
						prevStartTime = nil
					} else if lastClassEntry != nil { // Если же какое-то занятие уже парсится
						// Дополним в название
						lastClassEntry.Title += secondSubLine
						// Время начала теперь не нужно
						prevStartTime = nil
					} else if timeContainerRE.MatchString(firstSubLine) { // Если-же в первой подстроке есть какое-то время, то
						// Установим время начала
						f := utils.AddTimeToDate(date, timeContainerRE.FindStringSubmatch(firstLine)[1])
						prevStartTime = &f
					} else if timeContainerRE.MatchString(secondSubLine) { // Если-же во второй подстроке есть какое-то время, то
						// Установим время начала
						f := utils.AddTimeToDate(date, timeContainerRE.FindStringSubmatch(secondLine)[1])
						prevStartTime = &f
					}
				} else { // Если не занятие (а может и занятие)
					// Если цель дня уже есть, добавим разделитель

					// TODO: Нахуй колледж пишет со сколько пара? Или если мы не вышка, мы не догадаемся посмотреть направо?
					if crapyActivityHackRE.MatchString(firstSubLine) || crapyActivityHackRE.MatchString(secondSubLine) ||
						crapyActivityHack2RE.MatchString(firstSubLine) || crapyActivityHack2RE.MatchString(secondSubLine) {
						continue
					}

					if activity != "" {
						activity += " "
					}

					// Если время не указано, то добавим)
					if !timeIsUselessRE.MatchString(firstSubLine) && !timeIsUselessRE.MatchString(secondSubLine) && !timeIsUselessRE.MatchString(activity) {
						possibleActivityStartTime = startTime.Format("15:04")
					} else {
						possibleActivityStartTime = ""
					}

					// Добавим к цели подстроку
					activity += firstSubLine
					// И если есть, вторая под-строка
					if len(strings.TrimSpace(secondSubLine)) > 0 {
						// Добавим и её
						activity += ", " + secondSubLine
					}

				}
			}

			if lastClassEntry != nil {
				// Очистим занятие от повторяющихся пробелов
				lastClassEntry.Title = utils.RemoveSpaces(lastClassEntry.Title)

				// Дополним занятие в список занятий
				classes = append(classes, *lastClassEntry)
			}

			// Все, парсинг занятия закончился, можно обнулить
			lastClassEntry = nil
		}
	}

	// Если занятий нет и если нет цели дня, ничего не возвращаем
	if len(classes) == 0 && activity == "" {
		return nil, nil
	}

	if activity != "" && possibleActivityStartTime != "" {
		activity = "(" + possibleActivityStartTime + ") " + activity
	}

	// Запишем в Calendar[date] спарсеный с горечью и слезами день
	for _, name := range groups {
		classesHack := classes
		// Рубрика еженедельного костыля в фонд для колледжа
		if group.faculty == "Колледж" {
			classesHack = []model.Class{}
			for _, class := range classes {
				if !strings.Contains(class.Title, "(9)") && !strings.Contains(class.Title, "(11)") && //Лебедь, рак и щука. Спектакль от команды, состовляющей расписание
					!strings.Contains(class.Title, "/9") && !strings.Contains(class.Title, "/11") {
					classesHack = append(classesHack, class)
				} else {
					if strings.Contains(class.Title,
						strings.ReplaceAll(strings.ReplaceAll(name, "-9", ""), "-11", "")) {
						classesHack = append(classesHack, class)
					}
				}
			}

		}

		ret = append(ret, &model.Timetable{
			Institution: p.GetInstitution(),
			GroupName:   name,
			Faculty:     group.faculty,
			Date:        date,
			Activity:    activity,
			Classes:     classesHack,
		})
	}

	return nil, ret
}

// Иногда в расписание засовывают ссылку на занятие в зуме...
var externalLinkRE = regexp.MustCompile("(?i)(https?://[-a-zA-Zа-яё0-9+&@#/%?=~_|!:,{}.;]*[-a-zA-Z0-9+&@#/%=~_|])")

// Разобъем вторую строку по переносам строки и пройдемся по ним
func splitSubgroups(line string) (subgroups []string) {
	var subLineSplited []string
	for _, s := range strings.Split(line, "\n") {
		if len(s) > 0 {
			subLineSplited = append(subLineSplited, s)
		}
	}
	for lineIndex := 0; lineIndex < len(subLineSplited); lineIndex++ {
		// Уберём ссылку из подстроки
		cleaned := externalLinkRE.ReplaceAllString(subLineSplited[lineIndex], "")
		// Если это первая подстрока, либо в ней есть /
		if lineIndex == 0 || strings.Contains(cleaned, "/") {
			// Засунем в subgroups подгруппу
			subgroups = append(subgroups, subLineSplited[lineIndex])
		} else {
			// Иначе подгрупп нет (одна)
			subgroups = []string{strings.Join(subLineSplited, " ")}
			break
		}
	}
	return subgroups
}

// Обычно в первой строчке указано (лк), (пр), или (лб).(иногда в комбинациях)
var lectureRE = regexp.MustCompile("(?i)(лк|пр|лб)?(([, \\\\/|+]+)?(лк|пр|лб))+")

// Но не всегда. МДК, зачеты, экзамены и прочее говно иногда стоит без всего
var lectureBypassRE = regexp.MustCompile("(?i)(мдк|консультация|зач[её]т|математ|пересдача|экз|практи|психолог|социолог|тест|есстество|семьеведе|истори|защита|курсов|общество|иностранный|философии|ректорский|основы безопасности|экология|информатика|русский язык|литература|патриотическое|физическая|география|страховое|экономика|итоговая)")

// Занятие ли это
func isLecture(firstLine, secondLine string) bool {
	return lectureRE.MatchString(firstLine) || lectureBypassRE.MatchString(secondLine) ||
		campusName(secondLine) != emptyField
}

// Я скоро заебусь это поддерживать блять. А сейчас только 3 сентября нахуй. Допереворачивался календарь.
var anotherOneCollegeDirtyHack = regexp.MustCompile("^(.*) ([А-Яё][а-яё]+ +[А-ЯЁ][а-яё]+ +[А-ЯЁ][а-яё]+.*)$")

func resolveSubgroups(firstLineSplited, secondLineSplited []string, line int) [][]string {
	// Получим грубо говоря "нормальные" firstLine и secondLine
	firstSubLineRaw := strings.TrimSpace(utils.GetOrString(firstLineSplited, line, ""))
	secondSubLineRaw := strings.TrimSpace(utils.GetOrString(secondLineSplited, line, ""))
	if secondSubLineRaw == "" && line > 0 {
		secondSubLineRaw = strings.TrimSpace(utils.GetOrString(secondLineSplited, line-1, ""))
	}

	fSubgroups := splitSubgroups(firstSubLineRaw)
	sSubgroups := splitSubgroups(secondSubLineRaw)
	// college be like
	// Иностранный язык 2-ФИН (9)Б Пиксендеева Виктория Геннадьевна,  ауд., ул.Егорова, 16                                                         //  Иностранный язык 2-ФИН(9)Д+ 1-ФИН(11) Бажанская Инна Валентиновна, ауд. ,
	if len(fSubgroups) == 1 && len(sSubgroups) == 0 {
		if anotherOneCollegeDirtyHack.MatchString(fSubgroups[0]) {
			groups := anotherOneCollegeDirtyHack.FindStringSubmatch(fSubgroups[0])
			fSubgroups = []string{groups[1]}
			sSubgroups = []string{groups[2]}
		}
	}

	subgroups := make([][]string, (int)(math.Max(float64(len(fSubgroups)), float64(len(sSubgroups)))))
	for i := 0; i < len(subgroups); i++ {
		subgroups[i] = make([]string, 2)

		subgroups[i][0] = strings.TrimSpace(utils.GetOrString(fSubgroups, i, ""))
		subgroups[i][1] = strings.TrimSpace(utils.GetOrString(sSubgroups, i, ""))
	}

	if len(subgroups) > 1 {
		fmt.Println(subgroups)
	}

	return subgroups
}

// Нужно разбить последнюю строчку иногда по / (п/г и ссылки не подходят)
var subgroupBypassRE = regexp.MustCompile("(?i)([^п]|^https?:/)/([ ^г])?")

// Разобъем подстроку по /
func splitLine(line string) (spliced []string) {
	for _, tmp := range splitExcept(line, subgroupBypassRE, externalLinkRE) {
		tmp = strings.Trim(tmp, " /")
		if len(tmp) > 0 {
			spliced = append(spliced, tmp)
		}
	}
	// Колледж...
	if len(spliced) == 1 && lecturerNameCollegeRE.MatchString(line) {
		spliced = []string{}
		//TODO: Либо избавиться от костыля, либо слезно умолять раздел расписания вернуть все как было
		for _, tmp := range strings.SplitN(line, ",", 2) {
			tmp = strings.Trim(tmp, " /")
			if len(tmp) > 0 {
				spliced = append(spliced, tmp)
			}
		}
	}
	return spliced
}

// Очистка имени группы из вб от всякого говна
// Колледж: <курс>-<направление>[-<после какого класса>][+<совмещение>][[-](<заочно/очно/кабинет>)]
// Например: 1-ПКС-11, 1-ФИН-9+Д, 3-ЗИО
// Вышка: <курс><направление>[-<профиль>][[-](<заочно/очно/кабинет?/доктора?>)]
// Например: 1БПМИ-ПТ, 4БЛВ-ПРВ(403), 4БПО-НО-(з), 2СЛД(д)
var cleanRE = regexp.MustCompile("(?i)([0-9]-?[-а-яё0-9]+(\\+Д)? *(\\([0-9дз+]+\\))?)")

func cleanName(group string) (ret []string) {
	for _, name := range strings.Split(group, ",") {
		ret = append(ret, strings.TrimSpace(cleanRE.FindString(name)))
	}
	return ret
}

// Имя преподавателя в общем формате (В.Н. Морозов)
var lecturerNameCommonRE = regexp.MustCompile("(?i)([а-яё]\\.[а-яё])\\.? *([а-яё]+)")

// Имя преподавателя в формате, который частенько используется в колледже (Морозов Владислав Николаевич)
var lecturerNameCollegeRE = regexp.MustCompile("([А-Яё][а-яё]+) +([А-ЯЁ])[а-яё]+ +([А-ЯЁ])[а-яё]+")

// Имя преподавателя в формате Морозов В.Н
var lecturerNameInsaneRE = regexp.MustCompile("(?i)([а-яё]+) ([а-яё]\\.[а-яё])")

// Получить имя преподавателя в нужном формате по name
func lecturerName(name string) string {
	ret := name

	var k []string

	if k = lecturerNameCommonRE.FindStringSubmatch(name); len(k) != 0 {
		ret = k[1] + "." + k[2]
	} else if k = lecturerNameCollegeRE.FindStringSubmatch(name); len(k) != 0 {
		ret = strings.TrimSpace(k[2]) + "." + strings.TrimSpace(k[3]) + "." + strings.TrimSpace(k[1])
	} else if k = lecturerNameInsaneRE.FindStringSubmatch(name); len(k) != 0 {
		ret = strings.TrimSpace(k[2]) + "." + strings.TrimSpace(k[1])
	}

	if len(k) == 0 {
		return emptyField
	}

	//Удаляем слеши
	ret = strings.Replace(ret, "/", "", -1)
	//God save the queen
	ret = strings.Replace(ret, "Королева", "Королёва", -1)
	//Уберем крайние пробелы, табуляцию и переносы
	ret = strings.TrimSpace(ret)
	//Уберем пробелы
	ret = strings.Replace(ret, " ", "", -1)
	//Уберем точку в конце...
	ret = strings.TrimRight(ret, ".")

	return ret
}

// Все числа из строки, чтобы можно было получить аудиторию или корпус (это хак, но иначне никак)
var numberRE = regexp.MustCompile("(\\d+)")

// При учебной практике, во второй строчке прячутся невидимые даты. Поможем же Даше и Башмачку их найти
var practiceRE = regexp.MustCompile("\\d{2}\\.\\d{2}(\\.\\d{2,4})?[ ]*-[ ]*\\d{2}\\.\\d{2}(\\.\\d{2,4})?")

// Получить название корпуса по name
func campusName(name string) string {
	if practiceRE.MatchString(name) {
		return emptyField
	}
	if externalLinkRE.MatchString(name) {
		return name
	}
	var groups = numberRE.FindAllString(name, -1)
	var ret = name
	for i, auditorium := range groups {
		if auditorium == "57" {
			ret = "пр. Ленина 57, " + getExcept(groups, i)
		} else if auditorium == "9" {
			ret = "ул. Коммуны 9, " + getExcept(groups, i)
		} else if auditorium == "15" {
			ret = "ул. Капитана Егорова 15, " + getExcept(groups, i)
		} else if auditorium == "16" {
			ret = "ул. Капитана Егорова 16, " + getExcept(groups, i)
		}
	}

	return ret
}

// Извлечь и join'уть всё из groups (кроме i-того элемента)
func getExcept(groups []string, i int) (sb string) {
	for j, auditorium := range groups {
		if j != i {
			if sb != "" {
				sb += ", "
			}
			sb += auditorium
		}
	}
	return sb
}

// Модель для общения
type currentGroup struct {
	groupNames []string
	faculty    string
	sh         *xlsx.Sheet
}
