// parcel_test.go

package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func setupIntegrationTestDB() (*sql.DB, error) {
	return sql.Open("sqlite", "tracker.db")
}

func getParcelStore(t *testing.T) (ParcelStore, *sql.DB) {
	db, err := setupIntegrationTestDB()
	require.NoError(t, err)
	return NewParcelStore(db), db
}

func getTestParcel() Parcel {
	return Parcel{
		Client:    1,
		Status:    ParcelStatusRegistered,
		Address:   "Тестовый адрес",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func cleanupParcel(db *sql.DB, id int) {
	db.Exec("DELETE FROM parcel WHERE number = ?", id)
}

func TestAddGetDelete(t *testing.T) {
	store, db := getParcelStore(t)
	defer db.Close()

	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка добавления посылки")
	require.NotZero(t, id, "Идентификатор посылки должен быть больше нуля")
	defer cleanupParcel(db, id)

	storedParcel, err := store.Get(id)
	require.NoError(t, err, "Ошибка получения посылки")
	require.Equal(t, parcel.Client, storedParcel.Client, "Идентификатор клиента не совпадает")
	require.Equal(t, parcel.Status, storedParcel.Status, "Статус посылки не совпадает")
	require.Equal(t, parcel.Address, storedParcel.Address, "Адрес посылки не совпадает")

	err = store.Delete(id)
	require.NoError(t, err, "Ошибка удаления посылки")

	_, err = store.Get(id)
	require.Error(t, err, "Удалённая посылка не должна быть найдена")
}

func TestSetAddress(t *testing.T) {
	store, db := getParcelStore(t)
	defer db.Close()

	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка добавления посылки")
	defer cleanupParcel(db, id)

	newAddress := "Новый тестовый адрес"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err, "Ошибка изменения адреса")

	storedParcel, err := store.Get(id)
	require.NoError(t, err, "Ошибка получения посылки после изменения адреса")
	require.Equal(t, newAddress, storedParcel.Address, "Адрес посылки не обновился")
}

func TestSetStatus(t *testing.T) {
	store, db := getParcelStore(t)
	defer db.Close()

	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка добавления посылки")
	defer cleanupParcel(db, id)

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err, "Ошибка изменения статуса")

	storedParcel, err := store.Get(id)
	require.NoError(t, err, "Ошибка получения посылки после изменения статуса")
	require.Equal(t, ParcelStatusSent, storedParcel.Status, "Статус посылки не обновился")
}

func TestGetByClient(t *testing.T) {
	store, db := getParcelStore(t)
	defer db.Close()

	clientID := 123

	parcels := []Parcel{
		{Client: clientID, Status: ParcelStatusRegistered, Address: "Адрес 1", CreatedAt: "2024-01-01T00:00:00Z"},
		{Client: clientID, Status: ParcelStatusRegistered, Address: "Адрес 2", CreatedAt: "2024-01-02T00:00:00Z"},
	}
	parcelMap := make(map[int]Parcel)

	for _, p := range parcels {
		id, err := store.Add(p)
		require.NoError(t, err, "Ошибка добавления посылки")
		p.Number = id
		parcelMap[id] = p
	}
	defer func() {
		for id := range parcelMap {
			cleanupParcel(db, id)
		}
	}()

	// Получаем список посылок клиента
	storedParcels, err := store.GetByClient(clientID)
	require.NoError(t, err, "Ошибка получения посылок клиента")
	require.Len(t, storedParcels, len(parcels), "Количество посылок не совпадает")

	// Проверяем каждую полученную посылку
	for _, storedParcel := range storedParcels {
		originalParcel, exists := parcelMap[storedParcel.Number]
		require.True(t, exists, "Посылка с номером %d не найдена в ожидаемом списке", storedParcel.Number)

		require.Equal(t, originalParcel.Client, storedParcel.Client, "Идентификатор клиента не совпадает")
		require.Equal(t, originalParcel.Status, storedParcel.Status, "Статус посылки не совпадает")
		require.Equal(t, originalParcel.Address, storedParcel.Address, "Адрес посылки не совпадает")
	}
}

func TestSetAddress_InvalidStatus(t *testing.T) {
	store, db := getParcelStore(t)
	defer db.Close()

	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка добавления посылки")
	defer cleanupParcel(db, id)

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err, "Ошибка изменения статуса")

	err = store.SetAddress(id, "Новый тестовый адрес")
	require.Error(t, err, "Адрес не должен измениться для статуса 'отправлена'")
	require.Contains(t, err.Error(), "Изменение адреса невозможно")
}

func TestDelete_InvalidStatus(t *testing.T) {
	store, db := getParcelStore(t)
	defer db.Close()

	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка добавления посылки")
	defer cleanupParcel(db, id)

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err, "Ошибка изменения статуса")

	err = store.Delete(id)
	require.Error(t, err, "Удаление не должно быть доступно для статуса 'отправлена'")
	require.Contains(t, err.Error(), "Удаление невозможно")
}
