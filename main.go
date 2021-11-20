package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var numDaysInFuture int
var before string
var weekdaysOnly bool

func init() {
	rootCmd.PersistentFlags().IntVarP(&numDaysInFuture, "num-days-in-future", "n", 21, "config file (default is 21)")
	rootCmd.PersistentFlags().BoolVarP(&weekdaysOnly, "weekdays", "w", false, "omit weekends")
	rootCmd.PersistentFlags().StringVarP(&before, "before", "b", "", "omit start times after this time (format: 3:04 PM)")
}

func main() {

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {

	dir, err := homedir.Dir()
	if err != nil {
		return err
	}
	if err := godotenv.Load(path.Join(dir, ".env")); err != nil {
		return err
	}

	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "cale",
	Short: "cale is a Calendly helper",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		s := args[0]
		if strings.HasPrefix(s, "https") {
			return errors.New("URL argument not implemented yet")
		}
		slug := s

		token := os.Getenv("CALENDLY_API_KEY")

		url := url.URL{
			Scheme: "https",
			Host:   "api.calendly.com",
			Path:   "/event_types",
		}

		srcUrl := "https://calendly.com/api/v1/echo"
		srcUrl = "https://api.calendly.com/users/me"
		r, err := http.NewRequest(http.MethodGet, srcUrl, nil)
		if err != nil {
			return err
		}

		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var meResponse MeResponse
		if err := json.Unmarshal(data, &meResponse); err != nil {
			return err
		}
		userUUID := path.Base(meResponse.Resource.URI)

		uri := "https://api.calendly.com/users/" + userUUID
		q := url.Query()
		q.Set("user", uri)

		url.RawQuery = q.Encode()

		userReq, err := http.NewRequest(http.MethodGet, url.String(), nil)
		if err != nil {
			return err
		}

		userReq.Header.Set("Authorization", "Bearer "+token)
		userReq.Header.Set("Content-Type", "application/json")

		userResp, err := http.DefaultClient.Do(userReq)
		if err != nil {
			return err
		}

		if userResp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", userResp.StatusCode)
		}

		userData, err := ioutil.ReadAll(userResp.Body)
		if err != nil {
			return err
		}

		var v EventTypesResponse
		err = json.Unmarshal(userData, &v)
		if err != nil {
			return err
		}

		uuid, found := v.UUID(slug)
		if !found {
			return errors.New("slug not found")
		}

		rangeURL := RangeRequest{
			EventTypeUUID:          uuid,
			NumberOfDaysIntoFuture: numDaysInFuture,
		}.URL()

		resp, err = http.Get(rangeURL)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return errors.New("bad status code")
		}

		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var rangeResponse Response
		if err := json.Unmarshal(data, &rangeResponse); err != nil {
			return err
		}

		if len(rangeResponse.Days) == 0 {
			// TODO try bigger range
			return errors.New("no slots found")
		}

		for _, day := range rangeResponse.Days {
			for _, spot := range day.Spots {
				startTime, err := time.Parse(time.RFC3339, spot.StartTime)
				if err != nil {
					return err
				}
				if before != "" {
					timeOfDay, err := parseBefore(before)
					if err != nil {
						return err
					}
					y, m, d := startTime.Date()
					cutoff := time.Date(y, m, d, timeOfDay.Hour(), timeOfDay.Minute(), 0, 0, startTime.Location())
					if startTime.After(cutoff) {
						continue
					}
				}
				if weekdaysOnly && (startTime.Weekday() == time.Saturday || startTime.Weekday() == time.Sunday) {
					continue
				}
				date := startTime.Format("Mon 02 Jan\t03:04 PM")
				fmt.Println(date)
			}
		}
		return nil
	},
}

func parseBefore(s string) (time.Time, error) {
	formats := []string{
		"3",
		"3PM",
		"3 PM",
		"3:04PM",
		"3:04 PM",
	}
	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err != nil {
			continue
		}
		return t, nil
	}
	return time.Time{}, fmt.Errorf("could not parse time: %s", s)
}

type Response struct {
	Days []struct {
		Date   string `json:"date"`
		Status string `json:"status"`

		Spots []struct {
			Status    string `json:"status"`
			StartTime string `json:"start_time"`
		} `json:"spots"`
	} `json:"days"`
}

type RangeRequest struct {
	EventTypeUUID          string
	NumberOfDaysIntoFuture int
}

func (r RangeRequest) URL() string {
	q := url.Values{}
	q.Set("timezone", "America/New_York")
	q.Set("range_start", time.Now().Format("2006-01-02"))
	q.Set("range_end", time.Now().AddDate(0, 0, r.NumberOfDaysIntoFuture).Format("2006-01-02"))
	q.Set("diagnostics", "false")
	u := url.URL{
		Scheme:   "https",
		Host:     "calendly.com",
		Path:     fmt.Sprintf("/api/booking/event_types/%s/calendar/range", r.EventTypeUUID),
		RawQuery: q.Encode(),
	}
	return u.String()
}

type EventTypesResponse struct {
	Collection []struct {
		URI  string `json:"uri"`
		Slug string `json:"slug"`
	} `json:"collection"`
}

func (r *EventTypesResponse) UUID(slug string) (string, bool) {
	for _, v := range r.Collection {
		if v.Slug == slug {
			return path.Base(v.URI), true
		}
	}
	return "", false
}

type MeResponse struct {
	Resource struct {
		URI string `json:"uri"`
	} `json:"resource"`
}
