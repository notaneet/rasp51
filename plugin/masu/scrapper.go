package masu

import (
	"github.com/gocolly/colly"
	"github.com/tealeg/xlsx/v3"
	"io/ioutil"
	"net/http"
	"strings"
)

var timetableURL = "https://www.masu.edu.ru/student/timetable/"
var facultySelector = ".col-md-9 > ul:nth-child(2) > li > a[href]:nth-child(1)"
var timetableSelector = "body > div.main > div > div > div.col-md-9.col-md-pull-3.content > table > tbody > tr > td > a[href]"

func (p *_MASUMurmanskPlugin) scrap() error {
	c := colly.NewCollector(colly.Async(true))
	var err error
	c.OnHTML(facultySelector, func(e *colly.HTMLElement) {
		if _err := p.facultyScrapper(e); _err != nil {
			err = _err
		}
	})

	if _err := c.Visit(timetableURL); _err != nil {
		return _err
	}

	c.Wait()

	return err
}

func (p *_MASUMurmanskPlugin) facultyScrapper(e *colly.HTMLElement) error {
	link := e.Attr("href")
	faculty := facultyResolver(link)
	if !p.config.FacultyMatcher.Match(faculty) {
		//faculty не подходит
		return nil
	}

	c := colly.NewCollector(colly.Async(true))
	var err error = nil
	c.OnHTML(timetableSelector, func(e *colly.HTMLElement) {
		err = p.timetableScrapper(e, faculty)
	})
	if _err := c.Visit(e.Request.AbsoluteURL(link)); _err != nil {
		return _err
	}

	c.Wait()
	return err
}

func (p *_MASUMurmanskPlugin) timetableScrapper(e *colly.HTMLElement, faculty string) error {
	if strings.HasSuffix(e.Attr("href"), ".xlsx") {

		resp, err := http.Get(e.Request.AbsoluteURL(e.Attr("href")))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		f, err := xlsx.OpenBinary(b)
		if err != nil {
			return err
		}
		return p.parseWB(f, faculty)
	}
	return nil
}

func facultyResolver(link string) string {
	if strings.HasSuffix(link, "ikip/") {
		return "ИКИиП"
	} else if strings.HasSuffix(link, "ppi/") {
		return "ППИ"
	} else if strings.HasSuffix(link, "sgi/") {
		return "СГИ"
	} else if strings.HasSuffix(link, "fmen/") {
		return "МиЕН"
	} else if strings.HasSuffix(link, "fkbzhd/") {
		return "ФКиБДЖ"
	} else if strings.HasSuffix(link, "medical/") {
		return "Лечебное дело"
	} else if strings.HasSuffix(link, "college/") {
		return "Колледж"
	}

	return ""
}
