package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (suite *ParcelStoreTestSuite) TestSetAddress_InvalidStatus() {
	parcel := suite.getTestParcel()
	id, err := suite.store.Add(parcel)
	require.NoError(suite.T(), err, "Ошибка добавления посылки")
	defer suite.cleanupParcel(id)

	err = suite.store.SetStatus(id, ParcelStatusSent)
	require.NoError(suite.T(), err, "Ошибка изменения статуса")

	expErr := "Изменение адреса невозможно, посылка уже в пути"
	err = suite.store.SetAddress(id, "Новый тестовый адрес")
	require.Error(suite.T(), err, "Адрес не должен измениться для статуса 'отправлена'")
	assert.EqualError(suite.T(), err, expErr, "Изменение адреса невозможно")
}

func (suite *ParcelStoreTestSuite) TestDelete_InvalidStatus() {
	parcel := suite.getTestParcel()

	id, err := suite.store.Add(parcel)
	require.NoError(suite.T(), err, "Ошибка добавления посылки")
	defer suite.cleanupParcel(id)

	err = suite.store.SetStatus(id, ParcelStatusSent)
	require.NoError(suite.T(), err, "Ошибка изменения статуса")

	expErr := "Удаление невозможно, посылка уже в пути"
	err = suite.store.Delete(id)
	require.Error(suite.T(), err, "Удаление не должно быть доступно для статуса 'отправлена'")
	assert.EqualError(suite.T(), err, expErr, "Удаление невозможно")
}
