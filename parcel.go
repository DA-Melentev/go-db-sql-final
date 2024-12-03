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
		return 0, errors.New("Адрес не может быть пустым")
	}

	query := `INSERT INTO parcel (client, status, address, created_at) 
              VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(query, p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("Не удалось добавить посылку: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Не удалось получить идентификатор последней записи: %w", err)
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
		return p, fmt.Errorf("Посылка с номером %d не найдена", number)
	} else if err != nil {
		return p, fmt.Errorf("Не удалось получить данные посылки: %w", err)
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	query := `SELECT number, client, status, address, created_at 
              FROM parcel WHERE client = ?`
	rows, err := s.db.Query(query, client)
	if err != nil {
		return nil, fmt.Errorf("Не удалось получить посылки для клиента %d: %w", client, err)
	}
	defer rows.Close()

	var parcels []Parcel
	for rows.Next() {
		var p Parcel
		err = rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("Ошибка при обработке данных посылки: %w", err)
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
		return fmt.Errorf("Некорректный статус: %s", status)
	}

	query := `UPDATE parcel SET status = ? WHERE number = ?`
	_, err := s.db.Exec(query, status, number)
	if err != nil {
		return fmt.Errorf("Не удалось обновить статус: %w", err)
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	if address == "" {
		return errors.New("Адрес не может быть пустым")
	}

	query := `UPDATE parcel SET address = ? 
              WHERE number = ? AND status = ?`
	result, err := s.db.Exec(query, address, number, ParcelStatusRegistered)
	if err != nil {
		return fmt.Errorf("Не удалось обновить адрес: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Не удалось получить количество обновлённых строк: %w", err)
	}
	if affected == 0 {
		return errors.New("Изменение адреса невозможно, посылка уже в пути")
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {
	query := `DELETE FROM parcel WHERE number = ? AND status = ?`
	result, err := s.db.Exec(query, number, ParcelStatusRegistered)
	if err != nil {
		return fmt.Errorf("Не удалось удалить посылку: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Не удалось получить количество удалённых строк: %w", err)
	}
	if affected == 0 {
		return errors.New("Удаление невозможно, посылка уже в пути")
	}
	return nil
}
