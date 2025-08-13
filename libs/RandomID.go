package libs

import (
	uuid "github.com/google/uuid"
)

func GenerateRandomID(length int) string {
	id, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	return id.String()
}

func GenerateRandomIDWithPrefix(prefix string, length int) string {
	id, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	return prefix + id.String()
}

func GenerateRandomLengthNumbers(length int) uint64 {
	chars := "0123456789"
	result := make([]byte, length)
	for i := range length {
		result[i] = chars[uuid.New()[0]%10]
	}
	return uint64(result[0])
}

func GenerateUniqueID() string {
	id, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	return id.String()
}
