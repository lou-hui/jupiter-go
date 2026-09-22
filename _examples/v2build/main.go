package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	solanago "github.com/gagliardetto/solana-go"
	"github.com/lou-hui/jupiter-go/jupiter"
	"github.com/lou-hui/jupiter-go/jupiter/swapv2"
)

func main() {
	// Initialize the root client with API key (automatically added to all requests)
	apiKey := "jup_8b426080ae9d73d6138f9455eebd494368acd18ae6f9323fdf07cc00d82b332c"
	jupClient, err := jupiter.NewClient(
		jupiter.DefaultAPIURL,
		jupiter.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("x-api-key", apiKey)
			return nil
		}),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.TODO()

	// Call GET /swap/v2/build — combines quote + instructions in one request.
	// No prior /quote call needed: inputMint, outputMint, amount, and taker are
	// passed directly as query parameters.
	slippageBps := "50"
	buildResp, err := jupClient.BuildGetWithResponse(ctx, &swapv2.BuildGetParams{
		InputMint:   "So11111111111111111111111111111111111111112",
		OutputMint:  "JUPyiwrYJFskUPiHa7hkeR8VUtAeFoSYbKedZNsDvCN",
		Amount:      "100000",
		Taker:       "CijjYuhG92wQ9FFEpxKW8JqtzJEcwMqWmjLurtC3tsfx",
		SlippageBps: &slippageBps,
	})
	if err != nil {
		panic(err)
	}

	if buildResp.JSON200 == nil {
		panic("invalid BuildGetWithResponse response")
	}

	build := buildResp.JSON200
	if raw, err := json.MarshalIndent(build, "", "  "); err == nil {
		fmt.Println(string(raw))
	}
	fmt.Printf("Route: %s -> %s\n", *build.InputMint, *build.OutputMint)
	fmt.Printf("In: %s, Out: %s\n", *build.InAmount, *build.OutAmount)

	// Assemble a versioned transaction from the raw instructions returned by v2.
	// tx, err := buildTransaction(build, "CijjYuhG92wQ9FFEpxKW8JqtzJEcwMqWmjLurtC3tsfx")
	// if err != nil {
	// 	panic(err)
	// }

	// // Create a wallet from private key.
	// walletPrivateKey := "{YOUR_PRIVATE_KEY}"
	// wallet, err := solana.NewWalletFromPrivateKeyBase58(walletPrivateKey)
	// if err != nil {
	// 	panic(err)
	// }

	// // Create a Solana client. Change the URL to the desired Solana node.
	// solanaClient, err := solana.NewClient(wallet, "https://api.mainnet-beta.solana.com")
	// if err != nil {
	// 	panic(err)
	// }

	// // Sign and send the transaction.
	// signedTx, err := solanaClient.SendTransactionOnChain(ctx, tx)
	// if err != nil {
	// 	panic(err)
	// }

	// // Wait a bit to let the transaction propagate to the network.
	// // This is just an example and not a best practice.
	// time.Sleep(20 * time.Second)

	// // Check the transaction status.
	// _, err = solanaClient.CheckSignature(ctx, signedTx)
	// if err != nil {
	// 	panic(err)
	// }
}

// buildTransaction assembles a transaction from the raw instructions returned
// by GET /swap/v2/build. The blockhash embedded in blockhashWithMetadata is
// used directly so no extra RPC call is required.
func buildTransaction(build *swapv2.BuildResponse, feePayer string) (string, error) {
	feePayerKey, err := solanago.PublicKeyFromBase58(feePayer)
	if err != nil {
		return "", fmt.Errorf("invalid fee payer: %w", err)
	}

	var instructions []solanago.Instruction
	if build.ComputeBudgetInstructions != nil {
		for _, ix := range *build.ComputeBudgetInstructions {
			instructions = append(instructions, toSolanaInstruction(ix))
		}
	}
	if build.SetupInstructions != nil {
		for _, ix := range *build.SetupInstructions {
			instructions = append(instructions, toSolanaInstruction(ix))
		}
	}
	if build.SwapInstruction != nil {
		instructions = append(instructions, toSolanaInstruction(*build.SwapInstruction))
	}
	if build.CleanupInstruction != nil {
		instructions = append(instructions, toSolanaInstruction(*build.CleanupInstruction))
	}
	if build.TipInstruction != nil {
		instructions = append(instructions, toSolanaInstruction(*build.TipInstruction))
	}

	// Use the blockhash returned by the API directly — no extra RPC call needed.
	var recentBlockhash solanago.Hash
	if build.BlockhashWithMetadata != nil && build.BlockhashWithMetadata.Blockhash != nil {
		bh := *build.BlockhashWithMetadata.Blockhash
		for i := 0; i < 32 && i < len(bh); i++ {
			recentBlockhash[i] = byte(bh[i])
		}
	}

	solanaTx, err := solanago.NewTransaction(
		instructions,
		recentBlockhash,
		solanago.TransactionPayer(feePayerKey),
	)
	if err != nil {
		return "", fmt.Errorf("build transaction: %w", err)
	}

	txBytes, err := solanaTx.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("marshal transaction: %w", err)
	}

	return base64.StdEncoding.EncodeToString(txBytes), nil
}

func toSolanaInstruction(ix swapv2.Instruction) solanago.Instruction {
	programID := solanago.MustPublicKeyFromBase58(ix.ProgramId)

	var accounts []*solanago.AccountMeta
	for _, acc := range ix.Accounts {
		key := solanago.MustPublicKeyFromBase58(acc.Pubkey)
		accounts = append(accounts, &solanago.AccountMeta{
			PublicKey:  key,
			IsSigner:   acc.IsSigner,
			IsWritable: acc.IsWritable,
		})
	}

	data, _ := base64.StdEncoding.DecodeString(ix.Data)

	return &rawInstruction{
		programID: programID,
		accounts:  accounts,
		data:      data,
	}
}

type rawInstruction struct {
	programID solanago.PublicKey
	accounts  []*solanago.AccountMeta
	data      []byte
}

func (r *rawInstruction) ProgramID() solanago.PublicKey     { return r.programID }
func (r *rawInstruction) Accounts() []*solanago.AccountMeta { return r.accounts }
func (r *rawInstruction) Data() ([]byte, error)             { return r.data, nil }
