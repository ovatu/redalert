package web

import (
	"net/http"

	"github.com/ovatu/redalert/events"
	"github.com/ovatu/redalert/stats"
)

type checkPublic struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`

	Events []*events.Event        `json:"events"`
	Stats  stats.CheckStatsPublic `json:"stats"`
}

func statsHandler(c *appCtx, w http.ResponseWriter, r *http.Request) {

	checks := c.service.Checks()
	publicChecks := make([]checkPublic, len(checks))

	for idx, check := range checks {
		events, err := check.Store.GetRecent()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		publicChecks[idx] = checkPublic{
			ID:     check.Data.ID,
			Name:   check.Data.Name,
			Type:   check.Data.Type,
			Status: check.Data.Status.String(),
			Events: events,
			Stats:  check.Stats.Export(),
		}
	}

	Respond(w, publicChecks, 200)
}
