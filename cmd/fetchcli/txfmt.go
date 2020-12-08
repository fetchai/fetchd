package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	flagPort = "port"
)

type EncodeResponse struct {
	EncodedTx string `json:"encoded_tx"`
}

type EncodeErrResponse struct {
	Msg   string `json:"msg"`
	Error string `json:"error"`
}

func readStdTx(cdc *codec.Codec, stream io.Reader) (stdTx authtypes.StdTx, err error) {
	var bytes []byte
	bytes, err = ioutil.ReadAll(stream)

	if err != nil {
		return
	}

	if err = cdc.UnmarshalJSON(bytes, &stdTx); err != nil {
		return
	}

	return
}

func encodeJsonResponse(w http.ResponseWriter, resp interface{}, statusCode int) {
	encodedResp, err := json.Marshal(resp)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
	}
	w.Write(encodedResp)
}

func handleEncode(cdc *codec.Codec, w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		resp := EncodeErrResponse{
			Msg:   "POST only endpoint",
			Error: "",
		}
		encodeJsonResponse(w, resp, http.StatusMethodNotAllowed)
	}

	// read the transaction as posted into the http response
	stdTx, err := readStdTx(cdc, req.Body)
	if err != nil {
		resp := EncodeErrResponse{
			Msg:   "Unable to parse input transaction format",
			Error: err.Error(),
		}
		encodeJsonResponse(w, resp, http.StatusBadRequest)
		return
	}

	// re-encode it via the Amino wire protocol
	txBytes, err := cdc.MarshalBinaryLengthPrefixed(stdTx)
	if err != nil {
		resp := EncodeErrResponse{
			Msg:   "Unable to generate encoded transaction",
			Error: err.Error(),
		}
		encodeJsonResponse(w, resp, http.StatusInternalServerError)
		return
	}

	// base64 encode the encoded tx bytes (so that they can be passed directly to the tendermint protocol
	resp := EncodeResponse{
		EncodedTx: base64.StdEncoding.EncodeToString(txBytes),
	}
	encodeJsonResponse(w, resp, http.StatusOK)
}

func handleDecode(cdc *codec.Codec, w http.ResponseWriter, req *http.Request) {
	resp := EncodeErrResponse{
		Msg:   "Not implemented",
		Error: "",
	}
	encodeJsonResponse(w, resp, http.StatusNotImplemented)

}

func runTxFmtServer(cdc *codec.Codec) {
	bindPort := viper.GetInt(flagPort)
	bindAddr := fmt.Sprintf(":%d", bindPort)

	fmt.Printf("Listening on %s\n", bindAddr)

	http.HandleFunc("/encode", func(w http.ResponseWriter, req *http.Request) {
		handleEncode(cdc, w, req)
	})
	http.HandleFunc("/decode", func(w http.ResponseWriter, req *http.Request) {
		handleDecode(cdc, w, req)
	})
	http.ListenAndServe(bindAddr, nil)
}

func txFmtCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fmtd",
		Short: "Run the tx format server",
		RunE: func(cmd *cobra.Command, args []string) error {
			runTxFmtServer(cdc)
			return nil
		},
		Args: cobra.NoArgs,
	}

	cmd.Flags().Int(
		flagPort, 8090,
		"The port the server should run on",
	)

	return cmd
}
