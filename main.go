package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cli/cli/v2/pkg/cmd/factory"
	"github.com/cli/cli/v2/pkg/extensions"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
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

	now := time.Now()
	fromS := flag.String("from", now.AddDate(0, 0, -5).Format("2006-01-02"), "from date (YYYY-MM-DD) defaults to 5 days ago")
	toS := flag.String("to", now.Format("2006-01-02"), "to date (YYYY-MM-DD) defaults to today")
	flag.Parse()

	from, err := time.Parse("2006-01-02", lo.FromPtr(fromS))
	if err != nil {
		panic("invalid 'from' date format")
	}

	to, err := time.Parse("2006-01-02", lo.FromPtr(toS))
	if err != nil {
		panic("invalid 'to' date format")
	}

	if from.After(to) {
		panic("'from' date must be before 'to' date")
	}

	getContribResp, err := getContributions(response.Login, &from, &to)
	if err != nil {
		panic("error from github api: " + err.Error())
	}

	getContribResp.prettyPrint()
	checkVersion()
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

func (r GetContribResp) prettyPrint() error {
	total := r.
		User.
		ContributionsCollection.
		ContributionCalendar.
		TotalContributions

	var contribItems []ContributionDay
	for _, week := range r.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, day := range week.ContributionDays {
			contribItems = append(contribItems, day)
		}
	}

	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.AppendHeader(table.Row{"date", "level", "count"})
	tw.AppendRows(lo.Reverse(lo.Map(contribItems, func(item ContributionDay, _ int) table.Row {
		return table.Row{item.Date, item.ContributionLevel, item.ContributionCount}
	})))
	tw.AppendFooter(table.Row{"", "total", total})
	tw.SetStyle(table.StyleColoredBlackOnGreenWhite)
	tw.Render()

	return nil
}

func checkVersion() {
	f := factory.New("")
	extMgr := f.ExtensionManager
	exts := extMgr.List()

	if contribExt, found := lo.Find(exts, func(ext extensions.Extension) bool {
		return ext.Name() == "contrib"
	}); found {
		current := contribExt.CurrentVersion()
		latest := contribExt.LatestVersion()

		if current != latest {
			color.White(fmt.Sprintf("your contrib extension is out of date: %s -> %s", current, latest))
			color.White("run `gh extension upgrade contrib` to update")
		}
	}
}
