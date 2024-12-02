package main

import (
	"database/sql"
	"errors"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	if p.Address == "" {
		return 0, errors.New("address cannot be empty")
	}

	query := `INSERT INTO parcel (client, status, address, created_at) 
              VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(query, p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("failed to add parcel: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}
	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	query := `SELECT number, client, status, address, created_at 
              FROM parcel WHERE number = ?`
	row := s.db.QueryRow(query, number)

	var p Parcel
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return p, fmt.Errorf("parcel with number %d not found", number)
	} else if err != nil {
		return p, fmt.Errorf("failed to retrieve parcel: %w", err)
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	query := `SELECT number, client, status, address, created_at 
              FROM parcel WHERE client = ?`
	rows, err := s.db.Query(query, client)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve parcels for client %d: %w", client, err)
	}
	defer rows.Close()

	var parcels []Parcel
	for rows.Next() {
		var p Parcel
		err = rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan parcel: %w", err)
		}
		parcels = append(parcels, p)
	}
	return parcels, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	validStatuses := map[string]bool{
		ParcelStatusRegistered: true,
		ParcelStatusSent:       true,
		ParcelStatusDelivered:  true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	query := `UPDATE parcel SET status = ? WHERE number = ?`
	_, err := s.db.Exec(query, status, number)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	if address == "" {
		return errors.New("address cannot be empty")
	}

	query := `UPDATE parcel SET address = ? 
              WHERE number = ? AND status = ?`
	result, err := s.db.Exec(query, address, number, ParcelStatusRegistered)
	if err != nil {
		return fmt.Errorf("failed to update address: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve rows affected: %w", err)
	}
	if affected == 0 {
		return errors.New("address update not allowed")
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {
	query := `DELETE FROM parcel WHERE number = ? AND status = ?`
	result, err := s.db.Exec(query, number, ParcelStatusRegistered)
	if err != nil {
		return fmt.Errorf("failed to delete parcel: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve rows affected: %w", err)
	}
	if affected == 0 {
		return errors.New("deletion not allowed")
	}
	return nil
}
