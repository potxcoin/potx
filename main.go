package main

import (
	"encoding/hex"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/wallet"
	"math/big"
	"strings"
	"time"
)

var (
	rootCmd = &cobra.Command{
		Use:   "POTX",
		Short: "p",
		Long:  `POTX mining`,
	}
)
var priv string
var rpc string
var max_gas uint64
var pri_gas float64

func main() {
	mintCmd.Flags().StringVarP(&priv, "priv", "v", "", "priv key")
	mintCmd.Flags().StringVarP(&rpc, "rpc", "r", "https://rpc.ankr.com/eth", "rpc")
	mintCmd.Flags().Uint64VarP(&max_gas, "max_gas", "m", 30, "max_gas")
	mintCmd.Flags().Float64VarP(&pri_gas, "pri_gas", "p", 1.0, "pri_gas")
	rootCmd.AddCommand(mintCmd)
	rootCmd.Execute()
}

var logo = `
	8888888b.   .d88888b. 88888888888 Y88b   d88P 
	888   Y88b d88P" "Y88b    888      Y88b d88P  
	888    888 888     888    888       Y88o88P   
	888   d88P 888     888    888        Y888P    
	8888888P"  888     888    888        d888b    
	888        888     888    888       d88888b   
	888        Y88b. .d88P    888      d88P Y88b  
	888         "Y88888P"     888     d88P   Y88b 
`

const MINBALANCE = 2000000000
const POTXADDR = "0x5487147D1cd072f5C59a35E4588005861B024eF3"

var mintCmd = &cobra.Command{
	Use:   "mine",
	Short: "m",
	Long:  `mine`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\033[1;32;32m%s\033[0m\n", logo)

		key, err := ethkey(priv)
		client, err := jsonrpc.NewClient(rpc)
		chainID, err := client.Eth().ChainID()
		if err != nil {
			fmt.Println("rpc error")
		}
		fmt.Println("USING ADDRESS: ", key.Address().String())

		balance, err := client.Eth().GetBalance(key.Address(), ethgo.Latest)

		if balance.Uint64() < MINBALANCE {
			fmt.Println("BALANCE TOO LOW, AT LEAST 0.002 ETHER")
			return
		}

		if err != nil {
			fmt.Println(err)
			panic("Connect to ethereum network error")
		}

		nonce, err := client.Eth().GetNonce(key.Address(), ethgo.Latest)
		if err != nil {
			fmt.Println("Get Nonce Error, Please Retry")
		}

		ToAddr := ethgo.HexToAddress(POTXADDR)
		sendBz, _ := hex.DecodeString("0x")
		maxGasFee := new(big.Int).Mul(new(big.Int).SetUint64(max_gas), new(big.Int).SetUint64(1000000000))
		maxPriorityFeeFloat, _ := big.NewFloat(0).Mul(big.NewFloat(pri_gas), big.NewFloat(0).SetUint64(1000000000)).Uint64()
		maxPriorityFee := new(big.Int).SetUint64(maxPriorityFeeFloat)
		for {
			txn := &ethgo.Transaction{
				To:                   &ToAddr,
				Value:                new(big.Int).SetUint64(0),
				Nonce:                nonce,
				Input:                sendBz,
				Gas:                  150000,
				From:                 key.Address(),
				MaxPriorityFeePerGas: maxPriorityFee,
				MaxFeePerGas:         maxGasFee,
				Type:                 ethgo.TransactionDynamicFee,
				ChainID:              chainID,
			}
			signer := wallet.NewEIP155Signer(chainID.Uint64())
			txn, err = signer.SignTx(txn, key)
			data, err := txn.MarshalRLPTo(nil)
			txhash, err := client.Eth().SendRawTransaction(data)
			if err != nil {
				continue
			}
			for {
				receipt, err := client.Eth().GetTransactionReceipt(txhash)
				if receipt == nil {
					time.Sleep(time.Second)
					continue
				}
				if err != nil {
					continue
				}
				if receipt.Status == 1 {
					fmt.Printf("Receipt Tx for nonce: %d \n", nonce)
					if len(receipt.Logs) > 0 {
						// Got Mint
						fmt.Printf("\033[1;32;41m%s\033[0m\n", "SUCCESS POTX MINTED.")
					}
				}
				break
			}
			nonce++
		}

	},
}

func ethkey(priStr string) (ethgo.Key, error) {
	privBz, err := hex.DecodeString(strings.TrimPrefix(priStr, "0x"))
	if err != nil {
		return nil, err
	}
	key, err := wallet.NewWalletFromPrivKey(privBz)
	return key, err
}
