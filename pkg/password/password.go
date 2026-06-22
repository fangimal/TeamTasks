package password

import "golang.org/x/crypto/bcrypt"

const bcryptCost = bcrypt.DefaultCost

func Hash(plainPassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcryptCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func Compare(hash string, plainPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainPassword)) == nil
}
