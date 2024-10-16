package app

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"io"
	"os"
)

func GenerateSHA256FromFile(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Create a new SHA-256 hash object
	hash := sha256.New()

	// Read the file in chunks and write it to the hash
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	// Get the final SHA-256 checksum as a byte slice
	hashSum := hash.Sum(nil)

	// Convert the byte slice to a hexadecimal string and return it
	return hex.EncodeToString(hashSum), nil
}

func GenerateSha256Hex(dataToVerify []byte) (actualHashHex string) {
	actualHash32 := sha256.Sum256(dataToVerify)
	actualHash := actualHash32[:]
	actualHashHex = hex.EncodeToString(actualHash)

	return actualHashHex
}

func VerifySha256(dataToVerify []byte, expectedSha256Hex *string) (isVerified bool, actualHashHex string, err error) {
	if expectedSha256Hex == nil {
		return true, "", nil
	}

	expectedHash, err := hex.DecodeString(*expectedSha256Hex)
	if err != nil {
		return false, "", err
	}

	//if len(expectedHash) != sha256.Size {
	//	return false, fmt.Errorf("provided hex value \"%v\" of expected sha256 hash for NetworkConf file does not have expected length of %v bytes", expectedSha256Hex, sha256.Size)
	//}

	actualHash32 := sha256.Sum256(dataToVerify)
	actualHash := actualHash32[:]
	//if !bytes.Equal(expectedHash, actualHash) {
	//	actualHashHex := hex.EncodeToString(actualHash)
	//	return false, fmt.Errorf("provided expected sha256 \"%v\" for NetworkConf file does NOT match actual sha256 hash %v of the file content", expectedSha256Hex, actualHashHex)
	//}

	isVerified = bytes.Equal(expectedHash, actualHash)
	if !isVerified {
		actualHashHex = hex.EncodeToString(actualHash)
	}
	return isVerified, actualHashHex, nil
}

func VerifyAddressPrefix(addr string, expectedPrefix string) error {
	prefix, _, err := bech32.DecodeAndConvert(addr)
	if err != nil {
		return err
	}

	if prefix != expectedPrefix {
		return fmt.Errorf("invalid address prefix: expected %s, got %s", expectedPrefix, prefix)
	}

	return nil
}
