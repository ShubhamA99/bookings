package dbrepo

import (
	"context"
	"errors"
	"time"

	"github.com/tsawler/bookings-app/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "insert into reservation (first_name,Committee_name,email,phone,start_date,end_date,room_id,created_at,updated_at) values (?,?,?,?,?,?,?,?,?) "

	result, err := m.DB.ExecContext(ctx, stmt,
		res.FirstName,
		res.CommitteeName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return 0, err
	}
	newID, err := result.LastInsertId()
	if err != nil {

		return 0, err
	}

	return int(newID), nil
}

func (m *postgresDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "insert into room_restriction (start_date,end_date,room_id,reservation_id,created_at,updated_at,restriction_id) values (?,?,?,?,?,?,?)"

	_, err := m.DB.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		time.Now(),
		time.Now(),
		r.RestrictionID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) SearchAvailabilityByRoomID(start, end time.Time, roomID int) (bool, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var numRows int
	query := "select count(id)from room_restriction where room_id=? and ? < end_date and ? > start_date"

	row := m.DB.QueryRowContext(ctx, query, roomID, start, end)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}

	if numRows == 0 {
		return true, nil
	}

	return false, nil
}

func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := "select r.id,r.room_name from rooms r where r.id not in (select room_id from room_restriction rr where ? < rr.end_date and ? > rr.start_date )"

	rows, err := m.DB.QueryContext(ctx, query, start, end)
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

func (m *postgresDBRepo) GetRoomByID(id int) (models.Room, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := "select id,room_name,created_at,updated_at from rooms where id =?"

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&room.ID,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err != nil {
		return room, err
	}
	return room, nil
}

func (m *postgresDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "select id,first_name,Committee_name,email,password,access_level,created_at,updated_at from users where id =? "

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.User

	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.CommitteeName,
		&u.Username,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return u, err
	}

	return u, nil
}

func (m *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "update set first_name=?,Committee_name=?,Username=?,access_level=?,updated_at=? from users where id =?"

	_, err := m.DB.ExecContext(ctx, query,
		u.FirstName,
		u.CommitteeName,
		u.Username,
		u.AccessLevel,
		time.Now(),
	)

	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, "select id,password from users where Username = ?", email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil

}

func (m *postgresDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := "select r.id,r.first_name,r.Committee_name,r.email,r.phone,r.start_date,r.end_date,r.room_id,r.created_at,r.updated_at,rm.id,rm.room_name from reservation r left join rooms rm on(r.room_id=rm.id) order by r.start_date ASC"

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Reservation
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.CommitteeName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Room.ID,
			&i.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

func (m *postgresDBRepo) AllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := "select r.id,r.first_name,r.Committee_name,r.email,r.phone,r.start_date,r.end_date,r.room_id,r.created_at,r.updated_at,r.processed,rm.id,rm.room_name from reservation r left join rooms rm on(r.room_id=rm.id) where processed = 0 order by r.start_date ASC"

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Reservation
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.CommitteeName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Processed,
			&i.Room.ID,
			&i.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

func (m *postgresDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res models.Reservation

	query := "select r.id,r.first_name,r.Committee_name,r.email,r.phone,r.start_date,r.end_date,r.room_id,r.created_at,r.updated_at,r.processed,rm.id,rm.room_name from reservation r left join rooms rm on (r.room_id=rm.id) where r.id=?"

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&res.ID,
		&res.FirstName,
		&res.CommitteeName,
		&res.Email,
		&res.Phone,
		&res.StartDate,
		&res.EndDate,
		&res.RoomID,
		&res.CreatedAt,
		&res.UpdatedAt,
		&res.Processed,
		&res.Room.ID,
		&res.Room.RoomName,
	)

	if err != nil {
		return res, err
	}

	return res, nil
}

func (m *postgresDBRepo) UpdateReservation(u models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "update reservation set first_name=?,Committee_name=?,email=?,phone=?,updated_at=? where id =?"

	_, err := m.DB.ExecContext(ctx, query,
		u.FirstName,
		u.CommitteeName,
		u.Email,
		u.Phone,
		time.Now(),
		u.ID,
	)

	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "delete from reservation where id =?"

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) UpdateProcessedForReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "update  reservation set processed =? where id =?"

	_, err := m.DB.ExecContext(ctx, query, processed, id)
	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) AllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := "select id , room_name,created_at,updated_at from rooms order by room_name"

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return rooms, err
	}

	defer rows.Close()

	for rows.Next() {
		var rm models.Room
		err := rows.Scan(
			&rm.ID,
			&rm.RoomName,
			&rm.CreatedAt,
			&rm.UpdatedAt,
		)
		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, rm)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}
	return rooms, nil
}

// func (m *postgresDBRepo) GetRestrictionsForRoomBydate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	var restrictions []models.RoomRestriction

// 	query := "select id,reservation_id,restriction_id,room_id,start_date,end_date from room_restriction where ? < end_date and ? >= start_date and room_id =?"

// 	rows, err := m.DB.QueryContext(ctx, query, start, end, roomID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer rows.Close()

// 	for rows.Next() {
// 		var r models.RoomRestriction
// 		err := rows.Scan(
// 			&r.ID,
// 			&r.ReservationID,
// 			&r.RestrictionID,
// 			&r.RoomID,
// 			&r.StartDate,
// 			&r.EndDate,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		restrictions = append(restrictions, r)
// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return restrictions, nil

// }

// GetRestrictionsForRoomByDate returns restrictions for a room by date range
func (m *postgresDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var restrictions []models.RoomRestriction

	query := `
		select id, coalesce(reservation_id, 0), restriction_id, room_id, start_date, end_date
		from room_restriction where ? < end_date and ? >= start_date
		and room_id = ?
`

	rows, err := m.DB.QueryContext(ctx, query, start, end, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.RoomRestriction
		err := rows.Scan(
			&r.ID,
			&r.ReservationID,
			&r.RestrictionID,
			&r.RoomID,
			&r.StartDate,
			&r.EndDate,
		)
		if err != nil {
			return nil, err
		}
		restrictions = append(restrictions, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return restrictions, nil
}
