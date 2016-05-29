package gpgme

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/proglottis/oyster"
	"github.com/proglottis/oyster/config"
	"github.com/proglottis/oyster/cryptofs"
	"github.com/sourcegraph/rwvfs"
)

const testCipherText = `-----BEGIN PGP MESSAGE-----
Version: GnuPG v1
hQEMAw4698hF4WUhAQf/SdkbF/zUE6YjBxscDurrUZunSnt87kipLXypSxTDIdgj
O9huAaQwBz4uAJf2DuEN/7iAFGhi/v45NTujrG+7ocfjM3m/A2T80g4RVF5kKXBr
pFFgH7bMRY6VdZt1GKI9izSO/uFkoKXG8M31tCX3hWntQUJ9p+n1avGpu3wo4Ru3
CJhpL+ChDzXuZv4IK81ahrixEz4fJH0vd0TbsHpTXx4WPkTGXelM0R9PwiY7TovZ
onGZUIplrfA1HUUbQfzExyFw3oAo1/almzD5CBIfq5EnU8Siy5BNulDXm0/44h8A
lOUy6xqx7ITnWMYYf4a1cFoW80Yk+x6SYQqbcqHFQdJIAVr00V4pPV4ppFcXgdw/
BxKERzDupyqS0tcfVFCYLRmvtQp7ceOS6jRW3geXPPIz1U/VYBvKlvFu4XTMCS6z
4qY4SzZlFEsU
=IQOA
-----END PGP MESSAGE-----`

func TestGpgMEFS_CreateEncrypted(t *testing.T) {
	fs := setupCryptoFS(t)
	plain, err := fs.CreateEncrypted("test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(plain, "test\n"); err != nil {
		t.Fatal(err)
	}
	if err := plain.Close(); err != nil {
		t.Fatal(err)
	}

	cipher, err := fs.Open("test")
	if err != nil {
		t.Fatal(err)
	}
	defer cipher.Close()
	data, err := ioutil.ReadAll(cipher)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Errorf("Expected encrypted data, got nothing")
	}
}

func TestGpgMEFS_OpenEncrypted(t *testing.T) {
	fs := setupCryptoFS(t)
	cipher, err := fs.Create("test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(cipher, testCipherText); err != nil {
		t.Fatal(err)
	}
	if err := cipher.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = fs.OpenEncrypted("test")
	if err != nil {
		t.Fatal(err)
	}
}

func setupCryptoFS(t testing.TB) cryptofs.CryptoFS {
	os.Setenv("GNUPGHOME", "../testdata/gpghome")
	fs := New(rwvfs.Map(map[string]string{}), config.New())
	fs.SetCallback(func() []byte { return []byte("password") })
	if err := oyster.InitRepo(fs, []string{"test@example.com"}); err != nil {
		t.Fatal(err)
	}
	return fs
}
