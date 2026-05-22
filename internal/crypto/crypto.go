package crypto


import (
    "encoding/base64"
	"crypto/rand"
    "crypto/subtle"
	"fmt"
    "strings"

    "golang.org/x/crypto/argon2"
)

type HashParams struct {
    Memory      uint32
    Iterations  uint32
    Parallelism uint8
    SaltLength  uint32
    KeyLength   uint32
}

var defaultParams = &HashParams{
    Memory:      64 * 1024, // 64 MB — match python utility memory_cost
    Iterations:  3,         // match python time_cost
    Parallelism: 2,         // match python parallelism
    SaltLength:  16,
    KeyLength:   32,
}

func HashPassword(password string) (string, error) {
    salt := make([]byte, defaultParams.SaltLength)
    if _, err := rand.Read(salt); err != nil {
        return "", err
    }

    hash := argon2.IDKey(
        []byte(password),
        salt,
        defaultParams.Iterations,
        defaultParams.Memory,
        defaultParams.Parallelism,
        defaultParams.KeyLength,
    )

    // Encode to PHC string format — same format argon2-cffi produces
    encoded := fmt.Sprintf(
        "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
        argon2.Version,
        defaultParams.Memory,
        defaultParams.Iterations,
        defaultParams.Parallelism,
        base64.RawStdEncoding.EncodeToString(salt),
        base64.RawStdEncoding.EncodeToString(hash),
    )
    return encoded, nil
}

func VerifyPassword(password, encoded string) (bool, error) {
    parts := strings.Split(encoded, "$")
    if len(parts) != 6 {
        return false, fmt.Errorf("invalid hash format")
    }

    var p HashParams
    _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
    if err != nil {
        return false, err
    }

    salt, err := base64.RawStdEncoding.DecodeString(parts[4])
    if err != nil {
        return false, err
    }
    storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
    if err != nil {
        return false, err
    }

    p.KeyLength = uint32(len(storedHash))
    computedHash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

    // Constant-time comparison to prevent timing attacks
    if subtle.ConstantTimeCompare(storedHash, computedHash) == 1 {
        return true, nil
    }
    return false, nil
}
