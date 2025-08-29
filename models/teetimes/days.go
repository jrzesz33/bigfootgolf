package teetimes

import (
	"time"
)

type ReservedDay struct {
	ID  string    `json:"id,omitempty"`
	Day time.Time `json:"day"`
	//Reservations map[int]Reservation
	Times []Reservation `json:"reservations"`
}

func NewReservedDay(day time.Time, _season Season) ReservedDay {
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
		_blockSetting := _season.GetTimeDetails(day, _firstTime)
		if _blockSetting != nil {
			reservations = append(reservations, NewReservation(nil, nil, _firstTime, int64(_slot), *_blockSetting))
			_slot++
		}
		_firstTime = _firstTime.Add(_season.Gap)
	}
	resDay.Times = reservations
	return resDay
}
