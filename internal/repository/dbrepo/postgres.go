package dbrepo

import (
	"context"
	"github.com/nambroa/lodging-bookings/internal/models"
	"time"
)

func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // gives the transaction a 3-second timeout.
	defer cancel()

	var newID int

	stmt := `insert into reservations (first_name, last_name, email, phone, start_date, end_date, room_id, 
                          created_at, updated_at)
                          values($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}
	return newID, nil
}

// InsertRoomRestriction inserts a room restriction into the database.
func (m *postgresDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // gives the transaction a 3-second timeout.
	defer cancel()

	stmt := `insert into room_restrictions (start_date, end_date, room_id, reservation_id, created_at, updated_at,
                               restriction_id)
                               values ($1, $2, $3, $4, $5, $6, $7)`

	_, err := m.DB.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		time.Now(),
		time.Now(),
		r.RestrictionID)

	if err != nil {
		return err
	}
	return nil
}

// SearchAvailabilityByDatesByRoomID returns true if availability exists for a specific roomID, and false otherwise.
func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // gives the transaction a 3-second timeout.
	defer cancel()

	var numRows int

	query := `
		select
			count(id)
		from
		    room_restrictions
		where
		    room_id = $1 and
		    $2 > start_date and $3 < end_date;`

	row := m.DB.QueryRowContext(ctx, query, roomID, end, start)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}

	if numRows == 0 { // No matches in select query means date range is available for reservation.
		return true, nil
	}

	return false, nil
}

// SearchAvailabilityForAllRooms returns a slice of available rooms for a given date range.
func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // gives the transaction a 3-second timeout.
	defer cancel()

	var rooms []models.Room
	query := `select
				r.id, r.room_name
			  from
			      rooms r
			  where
			      r.id not in (select rr.room_id from room_restrictions rr where rr.start_date < $1 and rr.end_date > $2)
	
`
	rows, err := m.DB.QueryContext(ctx, query, end, start)
	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var room models.Room
		err := rows.Scan(
			&room.ID,
			&room.RoomName,
		)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}
	if err = rows.Err(); err != nil {
		return rooms, err
	}
	return rooms, nil
}
