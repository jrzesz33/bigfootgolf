package teetimes

import (
	"birdsfoot/app/models/db"
	"time"
)

type ReservedDay struct {
	ID  string    `json:"id,omitempty"`
	Day time.Time `json:"day"`
	//Reservations map[int]Reservation
	Times []Reservation `json:"reservations"`
}

func NewReservedDay(day time.Time, _season Season, _reserved []Reservation) ReservedDay {
	var resDay ReservedDay
	resDay.Day = day

	//add the reservations
	_slot := 1
	var reservations []Reservation
	_firstTime := _season.FirstTeeTime
	for {
		if _firstTime.After(_season.LastTeeTime) {
			break
		}
		reserved := checkIfReserved(int64(_slot), _reserved)
		if reserved != nil {
			reservations = append(reservations, *reserved)
			_slot++
		} else {
			_blockSetting := _season.GetTimeDetails(day, _firstTime)
			if _blockSetting != nil {
				_teeTime := time.Date(day.Year(), day.Month(), day.Day(), _firstTime.Hour(), _firstTime.Minute(), 0, 0, db.TimeLocation)
				reservations = append(reservations, NewReservation(nil, nil, _teeTime, int64(_slot), *_blockSetting))
				_slot++
			}
		}
		_firstTime = _firstTime.Add(_season.Gap)
	}
	resDay.Times = reservations
	return resDay
}

func checkIfReserved(slot int64, reserved []Reservation) *Reservation {
	for _, res := range reserved {
		if res.Slot == slot {
			return &res
		}
	}
	return nil
}
