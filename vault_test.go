package vivoupdater_test

import (
	"os"
	"testing"

	"github.com/OIT-ADS-Web/vivoupdater"
)

func TestVaultLogin(t *testing.T) {
	// NOTE: these are set as globals
	vivoupdater.VaultEndpoint = os.Getenv("VAULT_ENDPOINT")
	vivoupdater.VaultRoleId = os.Getenv("VAULT_ROLE_ID")
	vivoupdater.VaultSecretId = os.Getenv("VAULT_SECRET_ID")

	vaultConfig := &vivoupdater.VaultConfig{
		Endpoint: vivoupdater.VaultEndpoint,
		RoleId:   vivoupdater.VaultRoleId,
		SecretId: vivoupdater.VaultSecretId,
		// e.g. without Token yet
	}

	err := vivoupdater.FetchToken(vaultConfig)
	if err != nil {
		t.Errorf("could not login to vault err=%s\n", err)
	}
	if err == nil && len(vaultConfig.Token) == 0 {
		t.Error("could not get token from vault - although no error\n")
	}
}

func TestVaultRead(t *testing.T) {
	// NOTE: these are set as globals
	vivoupdater.VaultEndpoint = os.Getenv("VAULT_ENDPOINT")
	vivoupdater.VaultRoleId = os.Getenv("VAULT_ROLE_ID")
	vivoupdater.VaultSecretId = os.Getenv("VAULT_SECRET_ID")
	vivoupdater.AppEnvironment = os.Getenv("APP_ENVIRONMENT")

	vaultConfig := &vivoupdater.VaultConfig{
		Endpoint: vivoupdater.VaultEndpoint,
		RoleId:   vivoupdater.VaultRoleId,
		SecretId: vivoupdater.VaultSecretId,
		// e.g. without Token yet
	}

	// don't actually need config kafka fully to test vault
	kafkaConfig := &vivoupdater.KafkaSubscriber{}

	vivoupdater.GetCertsFromVault(vivoupdater.AppEnvironment,
		vaultConfig, kafkaConfig)

	if len(kafkaConfig.ClientCert) == 0 {
		t.Error("could not configure kafka client cert from vault\n")
	}
	if len(kafkaConfig.ClientKey) == 0 {
		t.Error("could not configure kafka client key from vault\n")
	}
	if len(kafkaConfig.ServerCert) == 0 {
		t.Error("could not configure kafka server cert from vault\n")
	}

}
