package main

import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"sort"
	"testing"
	"time"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

type ParcelStoreTestSuite struct {
	suite.Suite
	store ParcelStore
	db    *sql.DB
}

func setupIntegrationTestDB() (*sql.DB, error) {
	// Здесь создается подключение к базе данных (например, SQLite)
	// В реальной ситуации замените это на вашу базу данных
	return sql.Open("sqlite", "tracker.db")
}

func (suite *ParcelStoreTestSuite) SetupSuite() {
	var err error
	suite.db, err = setupIntegrationTestDB()
	require.NoError(suite.T(), err, "Ошибка подключения к базе данных")
	suite.store = NewParcelStore(suite.db)
}

func (suite *ParcelStoreTestSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *ParcelStoreTestSuite) cleanupParcel(id int) {
	_, err := suite.db.Exec("DELETE FROM parcel WHERE number = ?", id)
	require.NoError(suite.T(), err, "Ошибка очистки базы данных после теста")
}

func (suite *ParcelStoreTestSuite) getTestParcel() Parcel {
	return Parcel{
		Client:    1,
		Status:    ParcelStatusRegistered,
		Address:   "Тестовый адрес",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func (suite *ParcelStoreTestSuite) TestAddGetDelete() {
	parcel := suite.getTestParcel()

	id, err := suite.store.Add(parcel)
	require.NoError(suite.T(), err, "Ошибка добавления посылки")
	require.NotZero(suite.T(), id, "Идентификатор посылки должен быть больше нуля")
	defer suite.cleanupParcel(id)
	parcel.Number = id

	storedParcel, err := suite.store.Get(id)
	require.NoError(suite.T(), err, "Ошибка получения посылки")
	assert.Equal(suite.T(), parcel, storedParcel, "Структуры не совпадают. Ожидаемая: %v , полученная: %v", parcel, storedParcel)

	err = suite.store.Delete(id)
	require.NoError(suite.T(), err, "Ошибка удаления посылки")

	errMsg := fmt.Sprintf("Посылка с номером %d не найдена", id)
	_, err = suite.store.Get(id)
	assert.EqualError(suite.T(), err, errMsg, "Удалённая посылка не должна быть найдена")
}

func (suite *ParcelStoreTestSuite) TestSetAddress() {
	parcel := suite.getTestParcel()

	id, err := suite.store.Add(parcel)
	require.NoError(suite.T(), err, "Ошибка добавления посылки")
	defer suite.cleanupParcel(id)

	newAddress := "Новый тестовый адрес"
	err = suite.store.SetAddress(id, newAddress)
	require.NoError(suite.T(), err, "Ошибка изменения адреса")

	storedParcel, err := suite.store.Get(id)
	require.NoError(suite.T(), err, "Ошибка получения посылки после изменения адреса")
	assert.Equal(suite.T(), newAddress, storedParcel.Address, "Адрес посылки не обновился")
}

func (suite *ParcelStoreTestSuite) TestSetStatus() {
	parcel := suite.getTestParcel()

	id, err := suite.store.Add(parcel)
	require.NoError(suite.T(), err, "Ошибка добавления посылки")
	defer suite.cleanupParcel(id)

	err = suite.store.SetStatus(id, ParcelStatusSent)
	require.NoError(suite.T(), err, "Ошибка изменения статуса")

	storedParcel, err := suite.store.Get(id)
	require.NoError(suite.T(), err, "Ошибка получения посылки после изменения статуса")
	assert.Equal(suite.T(), ParcelStatusSent, storedParcel.Status, "Статус посылки не обновился")
}

func (suite *ParcelStoreTestSuite) TestGetByClient() {
	parcels := []Parcel{
		suite.getTestParcel(),
		suite.getTestParcel(),
		suite.getTestParcel(),
	}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i, p := range parcels {
		id, err := suite.store.Add(p)
		require.NoError(suite.T(), err, "Ошибка добавления посылки")
		parcels[i].Number = id
	}
	defer func() {
		for id := range parcels {
			suite.cleanupParcel(id)
		}
	}()

	// Получаем список посылок клиента
	storedParcels, err := suite.store.GetByClient(client)
	require.NoError(suite.T(), err, "Ошибка получения посылок клиента")
	assert.Len(suite.T(), storedParcels, len(parcels), "Количество посылок не совпадает")

	// Сортируем срезы по треку
	sort.Slice(parcels, func(i, j int) bool {
		return parcels[i].Number < parcels[j].Number
	})
	sort.Slice(storedParcels, func(i, j int) bool {
		return storedParcels[i].Number < storedParcels[j].Number
	})

	//сравниваем одним ассертом
	assert.Equal(suite.T(), parcels, storedParcels, "Вернувшийся срез не совпадает с ожидаемым")
}

/*
	Знаю что следующие два теста должны быть на уровне тестов сервиса, а сам функционал на уровне сервиса.
	Функционал перенес в сервис, тесты оставил
*/

func TestParcelStoreTestSuite(t *testing.T) {
	suite.Run(t, new(ParcelStoreTestSuite))
}
