package admin

import (
	"bigfoot/golf/common/models/teetimes"
	"encoding/json"
	"net/http"
	"time"
)

func GetSeasons(w http.ResponseWriter, r *http.Request) {

	_seas, err := teetimes.GetSeasons(time.Now())
	if err != nil {
		http.Error(w, "Error with Server", http.StatusBadRequest)
		return
	}

	if len(_seas) == 0 {
		//no seasons loaded so Init a new Season
		_seas = teetimes.InitNewSeason(time.Now().Year())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(_seas)

}
