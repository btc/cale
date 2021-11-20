package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRange(t *testing.T) {
	url := RangeRequest{
		EventTypeUUID:          "566ef9af-a93b-4330-94b3-ad3766a1b516",
		NumberOfDaysIntoFuture: 14,
	}.URL()
	resp, err := http.Get(url)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	var v Response
	err = json.Unmarshal(data, &v)
	require.NoError(t, err)
	t.Logf("%+v", string(data))

	require.NotEmpty(t, v.Days)
	for _, day := range v.Days {
		for _, spot := range day.Spots {
			startTime, err := time.Parse(time.RFC3339, spot.StartTime)
			require.NoError(t, err)
			date := startTime.Format("Mon, 02 Jan\t03:04 PM")
			t.Log(date)
		}
	}
}

func TestAuth(t *testing.T) {

	t.SkipNow()

	token := "TODO"

	srcUrl := "https://calendly.com/api/v1/echo"
	srcUrl = "https://api.calendly.com/users/me"
	r, err := http.NewRequest(http.MethodGet, srcUrl, nil)
	require.NoError(t, err)
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(r)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, data)
	var v MeResponse
	err = json.Unmarshal(data, &v)
	require.NoError(t, err)
	userUUID := path.Base(v.Resource.URI)
	require.Equal(t, "DEFHJNZJIVJOQRHY", userUUID)
}

func TestEventTypes(t *testing.T) {

	t.SkipNow()

	token := "TODO"

	url := url.URL{
		Scheme: "https",
		Host:   "api.calendly.com",
		Path:   "/event_types",
	}

	uri := "https://api.calendly.com/users/DEFHJNZJIVJOQRHY"
	q := url.Query()
	q.Set("user", uri)

	url.RawQuery = q.Encode()

	r, err := http.NewRequest(http.MethodGet, url.String(), nil)
	require.NoError(t, err)
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(r)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var v EventTypesResponse
	err = json.Unmarshal(data, &v)
	require.NoError(t, err)
	uuid, found := v.UUID("2h")
	require.True(t, found)
	require.Equal(t, "566ef9af-a93b-4330-94b3-ad3766a1b516", uuid)
}
