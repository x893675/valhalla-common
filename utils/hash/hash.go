package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

func MD5Bytes(b []byte) string {
	sum := md5.Sum(b)
	return hex.EncodeToString(sum[:])
}

func MD5(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func Sha1(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func EncryptPasswordWithCost(password string, cost int) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func IsPasswordEncrypted(password string) bool {
	cost, _ := bcrypt.Cost([]byte(password))
	return cost > 0
}

func EncryptPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ComparePassword(password, encryptionPassword string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(encryptionPassword), []byte(password)); err != nil {
		return false
	}
	return true
}

// CalculateMapChecksum orders the map according to its key, and calculating the overall md5 of the values.
// It's expected to work with ConfigMap (map[string]string) and Secrets (map[string][]byte).
func CalculateMapChecksum(data any) string {
	switch t := data.(type) {
	case map[string]string:
		return calculateMapStringString(t)
	case map[string][]byte:
		return calculateMapStringByte(t)
	default:
		return ""
	}
}

func calculateMapStringString(data map[string]string) string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	var checksum string

	for _, key := range keys {
		checksum += data[key]
	}

	return MD5(checksum)
}

func calculateMapStringByte(data map[string][]byte) string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	var checksum string

	for _, key := range keys {
		checksum += string(data[key])
	}

	return MD5(checksum)
}

func Hash(s []byte) uint32 {
	h := fnv.New32a()
	_, _ = h.Write(s)
	return h.Sum32()
}

func HashWithPrefix(prefix string, s []byte) string {
	h := fnv.New32a()
	_, _ = h.Write(s)
	return fmt.Sprintf("%s:%d", prefix, h.Sum32())
}

func Hash2(s []byte) string {
	h := fnv.New32a()
	_, _ = h.Write(s)
	return strconv.FormatUint(uint64(h.Sum32()), 10)
}

func HashWithPrefix2(prefix string, s []byte) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(prefix))
	_, _ = h.Write(s)
	return strconv.FormatUint(uint64(h.Sum32()), 10)
}
