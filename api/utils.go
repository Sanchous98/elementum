package api

import (
	"fmt"
	"net/url"

	"github.com/Sanchous98/elementum/config"
	"github.com/Sanchous98/elementum/util"
	"github.com/Sanchous98/elementum/xbmc"
)

// type contextMenu []*contextMenuItem
//
// type contextMenuItem []string
//
// // contextMenuRequest ...
// type contextMenuRequest struct {
// }
//
// func makeContextMenu(r contextMenuRequest) *contextMenu {
// 	m := &contextMenu{}
//
// }

func filterListItems(l xbmc.ListItems) xbmc.ListItems {
	t := config.Get().TraktToken != ""

	ret := make(xbmc.ListItems, 0)
	for _, i := range l {
		if !i.TraktAuth || t {
			ret = append(ret, i)
		}
	}

	return ret
}

// URLForHTTP ...

// URLForXBMC ...
func URLForXBMC(pattern string, args ...interface{}) string {
	u, _ := url.Parse(fmt.Sprintf(pattern, args...))
	return "plugin://" + config.Get().Info.ID + u.String()
}

// URLQuery ...
func URLQuery(route string, query ...string) string {
	v := url.Values{}
	for i := 0; i < len(query); i += 2 {
		v.Add(query[i], query[i+1])
	}
	return route + "?" + v.Encode()
}

func contextPlayURL(f string, forced bool) string {
	action := "links"
	if config.Get().ChooseStreamAuto {
		action = "play"
	}
	if forced {
		action = "force" + action
	}

	return fmt.Sprintf(f, action)
}

func contextPlayOppositeURL(f string, forced bool) string {
	action := "links"
	if !config.Get().ChooseStreamAuto {
		action = "play"
	}
	if forced {
		action = "force" + action
	}

	return fmt.Sprintf(f, action)
}
