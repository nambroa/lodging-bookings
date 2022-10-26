package dbrepo

import (
	"context"
	"errors"
	"github.com/nambroa/lodging-bookings/internal/models"
	"golang.org/x/crypto/bcrypt"
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

// GetRoomByID gets a room matching the id given as parameter.
func (m *postgresDBRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // gives the transaction a 3-second timeout.
	defer cancel()

	var room models.Room

	query := `select id, room_name, created_at, updated_at from rooms where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&room.ID,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt)

	if err != nil {
		return room, err
	}

	return room, nil
}

// GetUserByID returns a user with the given id as a parameter.
func (m *postgresDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // gives the transaction a 3-second timeout.
	defer cancel()

	query := `select id, first_name, last_name, email, password, access_level, created_at, updated_at
		      from users where id = $1 `

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.User
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt)

	if err != nil {
		return u, err
	}

	return u, nil

}

// UpdateUser updates a user in the database.
func (m *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // gives the transaction a 3-second timeout.
	defer cancel()

	query := `update users set first_name=$1, last_name=$2, email=$3, access_level=$4, updated_at=$5`

	_, err := m.DB.ExecContext(ctx, query, u.FirstName, u.LastName, u.Email, u.AccessLevel, time.Now())

	if err != nil {
		return err
	}

	return nil
}

// Authenticate authenticates a user.
func (m *postgresDBRepo) Authenticate(email, userTypedPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // gives the transaction a 3-second timeout.
	defer cancel()

	var id int // Holds the id of the authenticated user
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, "select id, password from users where email=$1", email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	// At this point, the user is fetched from the DB so the email is correct, but the password is not yet verified.

	// Compare the hash of the password entered by the user to the hash of the password stored in the DB.
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(userTypedPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	}
	if err != nil {
		return 0, "", err
	}
	return id, hashedPassword, nil
}

// GetAllReservations returns a slice of all reservations.
func (m *postgresDBRepo) GetAllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // gives the transaction a 3-second timeout.
	defer cancel()

	var reservations []models.Reservation

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id,
		r.created_at, r.updated_at, r.processed, rm.id, rm.room_name
		from reservations r
		left join rooms rm on (r.room_id = rm.id)
		order by r.start_date asc
`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()
	for rows.Next() {
		var reserv models.Reservation
		err = rows.Scan(&reserv.ID,
			&reserv.FirstName,
			&reserv.LastName,
			&reserv.Email,
			&reserv.Phone,
			&reserv.StartDate,
			&reserv.EndDate,
			&reserv.RoomID,
			&reserv.CreatedAt,
			&reserv.UpdatedAt,
			&reserv.Processed,
			&reserv.Room.ID,
			&reserv.Room.RoomName)
		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, reserv)
	}
	if err = rows.Err(); err != nil {
		return reservations, err
	}
	return reservations, nil

}

// GetNewReservations returns a slice of all new (not processed) reservations.
func (m *postgresDBRepo) GetNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // gives the transaction a 3-second timeout.
	defer cancel()

	var reservations []models.Reservation

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id,
		r.created_at, r.updated_at, rm.id, rm.room_name
		from reservations r
		left join rooms rm on (r.room_id = rm.id)
		where r.processed = 0
		order by r.start_date asc
`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()
	for rows.Next() {
		var reserv models.Reservation
		err = rows.Scan(&reserv.ID,
			&reserv.FirstName,
			&reserv.LastName,
			&reserv.Email,
			&reserv.Phone,
			&reserv.StartDate,
			&reserv.EndDate,
			&reserv.RoomID,
			&reserv.CreatedAt,
			&reserv.UpdatedAt,
			&reserv.Room.ID,
			&reserv.Room.RoomName)
		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, reserv)
	}
	if err = rows.Err(); err != nil {
		return reservations, err
	}
	return reservations, nil

}
