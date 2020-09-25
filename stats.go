package hltvscrape

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

const baseURL = "https://www.hltv.org"

// ExtractStats is used on a HLTV stats page, it extracts all data into applicable struct(s)
func ExtractStats(statsPageURL string) (data MapData, err error) { // @TODO Work on the extract stats function.

	//c := colly.NewCollector() // Could look to add options here for optimization
	return MapData{}, nil //placeholder
}

// ExtractMatch is used on a HLTV matchPage. It extracts all data from the matchpage, then will call functions to extract data from each player, and each map.
func ExtractMatch(url string) (match MatchData, err error) {

	// Initilizate MatchData with the data we have from just the URL.
	match = MatchData{
		MatchURL: url,
		MatchID:  strings.Split(url, "/")[4], //Extracting match id from URl, must have https:// prefixed
	}

	// Now we start scraping into the Struct
	c := colly.NewCollector()
	// Extract Team Data for both teams via the 'team' div class(s)

	c.OnHTML(`.team`, func(e *colly.HTMLElement) {
		// Look for data of first team
		team0URL, exists := e.DOM.Find(".team1-gradient").Children().Attr("href") // Locates
		if exists {
			extractTeamURLData(team0URL, &match.Team0)
			res, score, err := extractWinner(e)
			parseErr(err)
			match.Team0SeriesScore = score
			if res == 1 {
				match.Winner = 1
			}

		} //@TODO grab more data in both statements..
		team1URL, exists2 := e.DOM.Find(".team2-gradient").Children().Attr("href") // Locates
		if exists2 {
			extractTeamURLData(team1URL, &match.Team1)
			res, score, err := extractWinner(e)
			parseErr(err)
			match.Team1SeriesScore = score
			if res == 1 {
				match.Winner = 2
			}
		}
		if exists || exists2 {
			err = fmt.Errorf("couldn't get all teamdata")
		}

		// Match.winner is set if we can find a 'won' div in their children.
		// If we cant find in either, match has to be a draw and defaults to that.

	})
	// For each map link avaialable, get the statspage URL.
	c.OnHTML(`.results-stats`, func(e *colly.HTMLElement) {
		match.MapLinks = append(match.MapLinks, "https://www.hltv.org"+e.Attr("href"))
	})

	// This function extracts the length of the match (best of x)
	// It also extracts the stage of the tournament it is in.
	c.OnHTML(`.padding.preformatted-text`, func(e *colly.HTMLElement) {
		texts := strings.Split(e.Text, "\n")
		// We extract Best of X in top row.
		// Empty str in middle of slice
		// Match Context in bottom row
		if match.BestOfType, err = strconv.Atoi(strings.Split(texts[0], " ")[2]); err != nil {
			log.Printf("Error in extracting best of type. Attempted to extract from %s text.\n", e.Text)
			match.BestOfType = 0
		}
		match.Stage = texts[2][2:]
	})

	// This function extracts the name of the event that the match is played as a part of.
	c.OnHTML(`.event.text-ellipsis`, func(e *colly.HTMLElement) {
		match.Event = e.Text
		match.EventID = extractID(e.ChildAttr("a", "href"))
	})

	// This function extracts the exact unix time of the estimated match start

	c.OnHTML(`.timeAndEvent`, func(e *colly.HTMLElement) {
		match.MatchTimeEpoch, err = strconv.Atoi(e.ChildAttr(".time", "data-unix"))
		if err != nil {
			log.Printf("Could not get time from match page.\n")
		}
	})

	if err != nil {
		return match, err
	}
	c.Visit(url)

	return match, nil
}

// extractWinner extracts data from a 'teamx-gradient' html element from match pages.
// winner returns a 1 (for Won) that element has a 'won' div inside it. If there is a 'lost' div winner is a 0 (for lost), if a 'tie' div is found, winner is set to 2
func extractWinner(e *colly.HTMLElement) (winner int8, score SeriesScore, err error) {
	s := e.DOM.Children().Find(".won")

	if len(s.Nodes) > 0 {
		// If we get here, there is a 'won' div in e's children.
		pscore, err := strconv.Atoi(s.Text())
		parseErr(err)
		score = SeriesScore(pscore)
		winner = 1
	} else if l := e.DOM.Children().Find(".lost"); len(l.Nodes) > 0 {
		// Getting here means no won div was found.
		// We now check for a 'lost' div, if that is not found, check for a
		// If we get here, there is a 'lost' div in e's children.
		winner = 0
		pscore, err := strconv.Atoi(l.Text())
		if err != nil {
			parseErr(err)
		}
		score = SeriesScore(pscore)
	} else if d := e.DOM.Children().Find(".tie"); len(d.Nodes) > 0 {
		winner = 2
		pscore, err := strconv.Atoi(s.Text())
		if err != nil {
			parseErr(err)
		}
		score = SeriesScore(pscore)
	} else {
		return 2, 0, fmt.Errorf("couldn't find a result in HTML")
	}
	return
}

func extractTeamURLData(url string, t *Team) {
	t.TeamURL = baseURL + url
	t.TeamID = extractID(url)
	t.Name = strings.Split(url, "/")[3]

}

// extracts ID from relative URL
func extractID(url string) (id string) {
	id = strings.Split(url, "/")[2]
	return
}

// @TODO abstract team data extraction into function. smilar to 'extractWinner'
// @TODO check that match has starting via extracted unix timestamp of match start.

func parseErr(err error) error {

	if err != nil {
		return fmt.Errorf("error in extracting data from HTML: %s", err)
	}
	return nil

}
