package teetimes

import (
	"time"
)

type ReservedDay struct {
	ID           string              `json:"id,omitempty"`
	Day          time.Time           `json:"day"`
	Reservations map[int]Reservation `json:"reservations"`
	Times        []Reservation
}

func NewReservedDay(day time.Time, _firstTime time.Time, _lastTime time.Time, gap time.Duration) ReservedDay {
	var resDay ReservedDay
	resDay.Day = day

	//add the reservations
	_slot := 1
	reservations := make(map[int]Reservation)
	for {
		if _firstTime.After(_lastTime) {
			break
		}
		reservations[_slot] = NewReservation(nil, nil, _firstTime, int64(_slot))
		_firstTime = _firstTime.Add(gap)
		_slot++
	}
	resDay.Reservations = reservations
	return resDay
}
