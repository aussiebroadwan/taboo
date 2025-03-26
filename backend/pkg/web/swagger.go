package web

import (
	"net/http"

	_ "github.com/aussiebroadwan/taboo/backend/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Taboo Game API
// @version         1.0
// @description     This API provides endpoints to fetch game results from the Taboo game backend, including the latest game, specific games by ID, and a range of games.
//
// @contact.name    Lachlan Cox
// @contact.url     https://github.com/aussiebroadwan/taboo/issues
//
// @license.name	MIT
// @license.url		https://opensource.org/licenses/MIT
//
// @host			taboo.tabdiscord.com
// @BasePath		/api
//
// @schemes			https
func RegisterSwagger(router *http.ServeMux) {
	router.Handle("/swagger/", httpSwagger.Handler())
}
