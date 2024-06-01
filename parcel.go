package main

import (
	"database/sql"
	"errors"
	"time"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	query := "INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)"
	result, err := s.db.Exec(query, p.Client, ParcelStatusRegistered, p.Address, time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	query := "SELECT number, client, status, address, created_at FROM parcel WHERE number = ?"
	row := s.db.QueryRow(query, number)

	var p Parcel
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, errors.New("parcel not found")
		}
		return p, err
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	query := "SELECT number, client, status, address, created_at FROM parcel WHERE client = ?"
	rows, err := s.db.Query(query, client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parcels []Parcel
	for rows.Next() {
		var p Parcel
		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			return nil, err
		}
		parcels = append(parcels, p)
	}
	return parcels, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	query := "UPDATE parcel SET status = ? WHERE number = ?"
	_, err := s.db.Exec(query, status, number)
	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	query := "UPDATE parcel SET address = ? WHERE number = ? AND status = ?"
	result, err := s.db.Exec(query, address, number, ParcelStatusRegistered)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("address can only be changed if the parcel is in 'registered' status")
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {
	query := "DELETE FROM parcel WHERE number = ? AND status = ?"
	result, err := s.db.Exec(query, number, ParcelStatusRegistered)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("parcel can only be deleted if it is in 'registered' status")
	}
	return nil
}
