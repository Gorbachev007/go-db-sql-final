package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func setupDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	_, err = db.Exec(`CREATE TABLE parcel (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER,
		status TEXT,
		address TEXT,
		created_at TEXT
	)`)
	require.NoError(t, err)
	return db
}

// normalizeTime converts the time string to UTC format for comparison
func normalizeTime(t *testing.T, timeStr string) string {
	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	require.NoError(t, err)
	return parsedTime.UTC().Format(time.RFC3339)
}

func normalizeParcel(t *testing.T, parcel Parcel) Parcel {
	parcel.CreatedAt = normalizeTime(t, parcel.CreatedAt)
	return parcel
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db := setupDB(t)
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	parcel.Number = id
	parcel = normalizeParcel(t, parcel)

	// get
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	storedParcel = normalizeParcel(t, storedParcel)
	assert.Equal(t, parcel, storedParcel)

	// delete
	err = store.Delete(id)
	require.NoError(t, err)

	// check delete
	_, err = store.Get(id)
	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db := setupDB(t)
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	parcel.Number = id

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// update the address in the original parcel for comparison
	parcel.Address = newAddress
	parcel = normalizeParcel(t, parcel)

	// check
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	storedParcel = normalizeParcel(t, storedParcel)
	assert.Equal(t, parcel, storedParcel)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db := setupDB(t)
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	parcel.Number = id

	// set status
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	// update the status in the original parcel for comparison
	parcel.Status = newStatus
	parcel = normalizeParcel(t, parcel)

	// check
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	storedParcel = normalizeParcel(t, storedParcel)
	assert.Equal(t, parcel, storedParcel)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db := setupDB(t)
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
	}

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)
		parcels[i].Number = id
		parcels[i] = normalizeParcel(t, parcels[i])
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	for i := range storedParcels {
		storedParcels[i] = normalizeParcel(t, storedParcels[i])
	}

	// check
	for _, parcel := range storedParcels {
		expectedParcel, exists := parcelMap[parcel.Number]
		require.True(t, exists)
		assert.Equal(t, expectedParcel, parcel)
	}
}
