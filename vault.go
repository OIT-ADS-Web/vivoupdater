package vivoupdater

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	ms "github.com/mitchellh/mapstructure"
)

type auth struct {
	ClientToken string `json:"client_token"`
	Accessor    string `json:"accessor"`
}

type approle struct {
	Auth auth `json:"auth"`
}

func FetchToken(config *VaultConfig) error {
	vaultAddress := config.Endpoint
	roleID := config.RoleId
	secretID := config.SecretId
	endpoint := fmt.Sprintf("%s%s", vaultAddress, "auth/approle/login")

	message := map[string]interface{}{
		"role_id":   roleID,
		"secret_id": secretID,
	}

	messageJson, err := json.Marshal(message)
	if err != nil {
		return err
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(messageJson))
	fmt.Printf("%v\n", resp)
	if err != nil {
		return err
	}

	result := approle{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&result)
	if decodeErr != nil {
		return decodeErr
	}

	// make sense to do this?

	token := result.Auth.ClientToken
	fmt.Printf("token=%v\n", token)
	config.Token = token
	return nil
}

type VaultData struct {
	Data map[string]string `json:"data"`
}

// this is rather convoluted
func FetchSecrets(config *VaultConfig, paths map[string]string,
	obj interface{}) error {
	client := &http.Client{}

	//{"kafka.clientKey": "apps/scholars/development/kafka/clientKey" }
	results := make(map[string]interface{})

	bases := make(map[string]bool)
	for _, path := range paths {
		idx := strings.LastIndex(path, "/")
		base := path[:idx]
		bases[base] = true
	}

	gathered := make(map[string]interface{})

	for base, _ := range bases {
		url := fmt.Sprintf("%ssecret/%s", config.Endpoint, base)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("X-Vault-Token", config.Token)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		data := VaultData{}
		if err = json.Unmarshal(body, &data); err != nil {
			fmt.Printf("Unable to parse response from vault. The response was %s from vault.\n", resp.Status)
			return err
		} else {
			for k, v := range data.Data {
				gathered[fmt.Sprintf("%s/%s", base, k)] = v
			}
		}
	}

	for key, path := range paths {
		// NOTE: mapping to property syntax (like java)
		// not supported - have to change to underscore
		results[strings.Replace(key, ".", "_", -1)] = gathered[path]
	}
	// e.g map[kafka.clientKey] = "--- BEGIN ---"

	fmt.Printf("results=%v\n", results)
	// decode after we have all the values ...
	msConfig := &ms.DecoderConfig{
		DecodeHook: ms.StringToSliceHookFunc(","),
		Result:     &obj,
	}
	decoder, err := ms.NewDecoder(msConfig)
	if err != nil {
		return err
	}
	if err := decoder.Decode(results); err != nil {
		return err
	}

	return nil
}

/*
func FetchValues(config VaultConfig, path string, obj interface{}) error {
	url := fmt.Sprintf("%s%s", config.Endpoint, path)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Vault-Token", config.Token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	m := make(map[string]interface{})
	if err = json.Unmarshal(body, &m); err != nil {
		fmt.Printf("Unable to parse response from vault. The response was %s from vault.\n", resp.Status)
		return err
	} else {
		config := &ms.DecoderConfig{
			DecodeHook: ms.StringToSliceHookFunc(","),
			Result:     &obj,
		}
		decoder, err := ms.NewDecoder(config)
		if err != nil {
			return err
		}
		if err := decoder.Decode(m["data"]); err != nil {
			return err
		}
	}
	return nil
}
*/
