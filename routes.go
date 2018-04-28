package progressbar201X

import (
	"net/http"

	"github.com/sqrthree/progressbar201X/internal/controller"
)

type route struct {
	url    string
	method string
	handle func(http.ResponseWriter, *http.Request)
}

// init route rules.
var Routes = []route{
	{"/", "GET", controller.Pong},
	{"/", "POST", controller.HandleEvents},
}
