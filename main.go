package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/table"
	"github.com/samber/lo"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			color.Red(fmt.Sprint(r))
		}
	}()

	client, err := api.DefaultRESTClient()
	if err != nil {
		panic(err)
	}

	response := struct{ Login string }{}
	if err := client.Get("user", &response); err != nil {
		panic("error from github api: " + err.Error())
	}

	to := time.Now()
	from := to.AddDate(0, 0, -5)

	getContribResp, err := getContributions(response.Login, &from, &to)
	if err != nil {
		panic("error from github api: " + err.Error())
	}

	total := getContribResp.
		User.
		ContributionsCollection.
		ContributionCalendar.
		TotalContributions

	var contribItems ContributionList
	for _, week := range getContribResp.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, day := range week.ContributionDays {
			contribItems = append(contribItems, day)
		}
	}

	fmt.Printf("total contributions from %s to %s: %d\n", from.Format("2006-01-02"), to.Format("2006-01-02"), total)
	contribItems.PrettyPrint()
}

type GetContribResp struct {
	User struct {
		ContributionsCollection struct {
			ContributionCalendar struct {
				TotalContributions int `json:"totalContributions"`
				Weeks              []struct {
					ContributionDays []ContributionDay `json:"contributionDays"`
				} `json:"weeks"`
			} `json:"contributionCalendar"`
		} `json:"contributionsCollection"`
	} `json:"user"`
}
type ContributionDay struct {
	ContributionCount int                  `json:"contributionCount"`
	ContributionLevel ContributionQuartile `json:"contributionLevel"`
	Date              string               `json:"date"`
}

type ContributionQuartile string

const (
	ContributionLevel1 ContributionQuartile = "FIRST_QUARTILE"
	ContributionLevel2 ContributionQuartile = "SECOND_QUARTILE"
	ContributionLevel3 ContributionQuartile = "THIRD_QUARTILE"
	ContributionLevel4 ContributionQuartile = "FOURTH_QUARTILE"
)

// getContributions returns the contributions for a user
func getContributions(userName string, from, to *time.Time) (*GetContribResp, error) {
	const query = `
query($userName:String!, $from: DateTime, $to: DateTime) {
  user(login: $userName){
    contributionsCollection(from: $from, to: $to) {
      contributionCalendar {
        totalContributions
        weeks {
          contributionDays {
            contributionCount
						contributionLevel
            date
          }
        }
      }
    }
  }
}`

	variables := map[string]any{
		"userName": userName,
	}
	if !lo.FromPtr(from).IsZero() {
		variables["from"] = from.Format(time.RFC3339)
	}
	if !lo.FromPtr(to).IsZero() {
		variables["to"] = to.Format(time.RFC3339)
	}

	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, err
	}

	var response GetContribResp
	if err := client.Do(query, variables, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

type ContributionList []ContributionDay

func (l ContributionList) PrettyPrint() error {
	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.AppendHeader(table.Row{"Date", "ContribCnt", "Level"})
	tw.AppendRows(lo.Map(l, func(item ContributionDay, _ int) table.Row {
		return table.Row{item.Date, item.ContributionCount, item.ContributionLevel}
	}))
	tw.SetStyle(table.StyleColoredBlackOnGreenWhite)
	tw.Render()

	return nil
}
