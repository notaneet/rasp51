package main

import (
	"flag"
	"fmt"
	"github.com/notaneet/rasp51/config"
	"github.com/notaneet/rasp51/converter"
	"github.com/notaneet/rasp51/plugin"
	"os"
)

func main() {
	var cfg = config.ParserConfig{}

	flag.Var(&cfg.GroupMatcher.MatchRaw, "group", "Требуемые группы")
	flag.Var(&cfg.FacultyMatcher.MatchRaw, "faculty", "Требуемые факультеты")
	flag.Var(&cfg.LecturerMatcher.MatchRaw, "lecturer", "Требуемые преподаватели")
	flag.Var(&cfg.ClassMatcher.MatchRaw, "class", "Требуемые предметы")
	flag.Var(&cfg.CampusMatcher.MatchRaw, "campus", "Требуемые корпуса и аудитории")
	flag.StringVar(&cfg.Interval, "interval", "", "Требуемое время [дд.мм.гггг[-дд.мм.гггг]]")

	var (
		output,
		pluginName,
		converterName string
	)

	flag.StringVar(&output, "output", "data.out", "Файл, куда будет записываться результат")
	flag.StringVar(&pluginName, "plugin", "МАГУ", "Требуемое учебное учреждение")
	flag.StringVar(&converterName, "converter", "pjson", "Тип выходных данных")

	flag.Parse()

	if pluginName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	cfg.Init()

	plug := plugin.NewPlugin(pluginName, cfg)
	if plug == nil {
		fmt.Println(pluginName + " не найден. ")
		os.Exit(1)
	}

	err, timetable := plug.GetTimetable()
	if err != nil {
		fmt.Println("Ошибка при парсинге расписания, ", err)
		os.Exit(1)
	}

	c := converter.Converter(converterName)
	err = c.Write(timetable, output)
	if err != nil {
		fmt.Println("Ошибка в сохранении расписания, ", err)
		os.Exit(1)
	}

}
