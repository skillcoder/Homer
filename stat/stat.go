package stat

import (
	"net/http"

	"github.com/skillcoder/homer/database"
	"github.com/takama/router"
)

type serviceStat struct {
	DB *database.StatT `json:"db"`
}

// Handler provides JSON API response giving service information
func Handler(dbGetStat func() database.StatT) router.Handle {
	return func(c *router.Control) {
		databaseStat := dbGetStat()

		info := serviceStat{
			DB: &databaseStat,
		}

		c.Code(http.StatusOK).Body(info)
	}
}
