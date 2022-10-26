package dbrepo

import (
	"errors"
	"github.com/nambroa/lodging-bookings/internal/models"
	"time"
)

func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	return 1, nil
}

// InsertRoomRestriction inserts a room restriction into the database.
func (m *testDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {

	if r.RoomID > 2 {
		return errors.New("non-existent room restriction test case")
	}

	return nil
}

// SearchAvailabilityByDatesByRoomID returns true if availability exists for a specific roomID, and false otherwise.
func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	return false, nil
}

// SearchAvailabilityForAllRooms returns a slice of available rooms for a given date range.
func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {

	var rooms []models.Room
	return rooms, nil
}

// GetRoomByID gets a room matching the id given as parameter.
func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	var room models.Room

	if id > 2 {
		return room, errors.New("non-existent room test case")
	}

	return room, nil
}

func (m *testDBRepo) GetUserByID(id int) (models.User, error)                  { return models.User{}, nil }
func (m *testDBRepo) UpdateUser(u models.User) error                           { return nil }
func (m *testDBRepo) Authenticate(email, password string) (int, string, error) { return 1, "", nil }
func (m *testDBRepo) GetAllReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}
func (m *testDBRepo) GetNewReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}
