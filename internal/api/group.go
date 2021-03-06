package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	gm "github.com/energieip/common-components-go/pkg/dgroup"
	"github.com/energieip/common-components-go/pkg/dserver"
	"github.com/energieip/common-components-go/pkg/duser"
	"github.com/energieip/common-components-go/pkg/tools"
	"github.com/energieip/srv200-coreservice-go/internal/database"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

func (api *API) readGroupConfig(w http.ResponseWriter, grID int) {
	group, _ := database.GetGroupConfig(api.db, grID)
	if group == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Group "+strconv.Itoa(grID)+" not found", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(group)
}

func (api *API) getGroupSetup(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	grID, err := strconv.Atoi(params["groupID"])
	if err != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Group "+strconv.Itoa(grID)+" not found", http.StatusInternalServerError)
		return
	}
	api.readGroupConfig(w, grID)
}

func (api *API) setGroupSetup(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	api.setGroupConfig(w, req)
}

func (api *API) setGroupConfig(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin, duser.PriviledgeMaintainer}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	gr := gm.GroupConfig{}
	err = json.Unmarshal(body, &gr)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	event := make(map[string]interface{})
	event["group"] = gr
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) sendGroupCommand(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	gr := dserver.GroupCmd{}
	err = json.Unmarshal(body, &gr)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	if api.hasEnoughRight(w, req, gr.Group) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	event := make(map[string]interface{})
	event["groupCmd"] = gr
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) removeGroupSetup(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	if api.hasAccessMode(w, req, []string{duser.PriviledgeAdmin}) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(req)
	grID := params["groupID"]
	i, err := strconv.Atoi(grID)
	if err != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Group "+grID+" not found", http.StatusInternalServerError)
		return
	}
	res := database.RemoveGroupConfig(api.db, i)
	if res != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Group "+grID+" not found", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("{}"))
}

func (api *API) getGroupStatus(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	params := mux.Vars(req)
	grID := params["groupID"]
	i, err := strconv.Atoi(grID)
	if err != nil {
		api.sendError(w, APIErrorDeviceNotFound, "Group "+grID+" not found", http.StatusInternalServerError)
		return
	}
	if api.hasEnoughRight(w, req, i) != nil {
		api.sendError(w, APIErrorUnauthorized, "Unauthorized Access", http.StatusUnauthorized)
		return
	}
	res := database.GetGroupStatus(api.db, i)
	if res == nil {
		api.sendError(w, APIErrorDeviceNotFound, "Group "+grID+" not found", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(res)
}

func (api *API) getGroupsStatus(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Connection", "close")
	defer req.Body.Close()
	decoded := context.Get(req, "decoded")
	var auth duser.UserAccess
	mapstructure.Decode(decoded.(duser.UserAccess), &auth)

	res := database.GetGroupsStatus(api.db)
	var groups []gm.GroupStatus
	for _, g := range res {
		if auth.Priviledge == duser.PriviledgeUser {
			if !tools.IntInSlice(g.Group, auth.AccessGroups) {
				continue
			}
		}
		groups = append(groups, g)
	}
	json.NewEncoder(w).Encode(groups)
}
