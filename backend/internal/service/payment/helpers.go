package payment

import (
	"LuomuTori/internal/config"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"gitlab.com/moneropay/go-monero/walletrpc"
	moneropay "gitlab.com/moneropay/moneropay/v2/pkg/model"
	"io"
	"net/http"
)

const DepositRoute = "/deposit"

func moneropayDepositCallbackURL(userID uuid.UUID) string {
	return "http://" + config.InternalAddr + DepositRoute + "?user-id=" + userID.String()
}

func moneropayReceivePost(amount uint64, description string, callbackURL string) (*moneropay.ReceivePostResponse, error) {
	req := moneropay.ReceivePostRequest{
		Amount:      amount,
		Description: description,
		CallbackUrl: callbackURL,
	}

	reqBytes, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}

	url := config.MoneropayURL + "/receive"

	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Invalid response: %v\n", resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	invoice := new(moneropay.ReceivePostResponse)
	if err := json.Unmarshal(body, invoice); err != nil {
		return nil, err
	}

	return invoice, nil
}

func moneropayReceiveGet(address string) (*moneropay.ReceiveGetResponse, error) {
	url := config.MoneropayURL + fmt.Sprintf("/receive/%s", address)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response: %v\n", resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data := new(moneropay.ReceiveGetResponse)
	if err := json.Unmarshal(body, data); err != nil {
		return nil, err
	}

	return data, nil
}

func moneropayTransferPost(destinations []walletrpc.Destination) (*moneropay.TransferPostResponse, error) {
	req := moneropay.TransferPostRequest{
		Destinations: destinations,
	}

	reqBytes, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}

	url := config.MoneropayURL + "/transfer"
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Invalid response: %v\n", resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data := new(moneropay.TransferPostResponse)
	if err := json.Unmarshal(body, data); err != nil {
		return nil, err
	}

	return data, nil
}

func moneropayTransferGet(txHash string) (*moneropay.TransferGetResponse, error) {
	url := config.MoneropayURL + fmt.Sprintf("/transfer/%s", txHash)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid response: %v\n", resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data := new(moneropay.TransferGetResponse)
	if err := json.Unmarshal(body, data); err != nil {
		return nil, err
	}

	return data, nil
}
