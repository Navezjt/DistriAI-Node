package distri

import (
	"DistriAI-Node/chain"
	"DistriAI-Node/chain/distri/distri_ai"
	"DistriAI-Node/docker"
	"DistriAI-Node/machine_info"
	"DistriAI-Node/pattern"
	"DistriAI-Node/utils"
	logs "DistriAI-Node/utils/log_utils"
	"context"
	"encoding/json"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type WrapperDistri struct {
	*chain.InfoChain
}

// Register the given hardware information with a distributed system or blockchain
func (chain WrapperDistri) AddMachine(hardwareInfo machine_info.MachineInfo) (string, error) {
	logs.Normal(fmt.Sprintf("Extrinsic : %v", pattern.TX_HASHRATE_MARKET_REGISTER))

	recent, err := chain.Conn.RpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("> GetRecentBlockhash: %v", err.Error())
	}

	uuid, err := utils.ParseMachineUUID(string(hardwareInfo.MachineUUID))
	if err != nil {
		return "", fmt.Errorf("> ParseMachineUUID: %v", err.Error())
	}

	jsonData, err := json.Marshal(hardwareInfo)
	if err != nil {
		return "", fmt.Errorf("> json.Marshal: %v", err.Error())
	}

	distri_ai.SetProgramID(chain.ProgramDistriID)

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			distri_ai.NewAddMachineInstruction(
				uuid,
				string(jsonData),
				chain.ProgramDistriMachine,
				chain.Wallet.Wallet.PublicKey(),
				solana.SystemProgramID,
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(chain.Wallet.Wallet.PublicKey()),
	)

	if err != nil {
		return "", fmt.Errorf("> NewAddMachineInstruction: %v", err.Error())
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if chain.Wallet.Wallet.PublicKey().Equals(key) {
				return &chain.Wallet.Wallet.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("> tx.Sign: %v", err.Error())
	}

	logs.Normal("=============== AddMachine Transaction")
	spew.Dump(tx)

	sig, err := chain.Conn.SendAndConfirmTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("> SendAndConfirmTransaction: %v", err.Error())
	}

	logs.Vital(fmt.Sprintf("%s completed : %v", pattern.TX_HASHRATE_MARKET_REGISTER, sig))

	return sig, nil
}

func (chain WrapperDistri) RemoveMachine() (string, error) {
	logs.Normal(fmt.Sprintf("Extrinsic : %s", pattern.TX_HASHRATE_MARKET_REMOVE_MACHINE))

	recent, err := chain.Conn.RpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	distri_ai.SetProgramID(chain.ProgramDistriID)
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			distri_ai.NewRemoveMachineInstruction(
				chain.ProgramDistriMachine,
				chain.Wallet.Wallet.PublicKey(),
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(chain.Wallet.Wallet.PublicKey()),
	)

	if err != nil {
		return "", fmt.Errorf("error creating transaction: %v", err)
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if chain.Wallet.Wallet.PublicKey().Equals(key) {
				return &chain.Wallet.Wallet.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("error signing transaction: %v", err)
	}

	logs.Normal("=============== RemoveMachine Transaction")
	spew.Dump(tx)

	sig, err := chain.Conn.SendAndConfirmTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("> SendAndConfirmTransaction: %v", err.Error())
	}

	logs.Vital(fmt.Sprintf("%s completed : %v", pattern.TX_HASHRATE_MARKET_REMOVE_MACHINE, sig))

	return sig, nil
}

func (chain WrapperDistri) OrderStart() (string, error) {
	logs.Normal(fmt.Sprintf("Extrinsic : %v", pattern.TX_HASHRATE_MARKET_ORDER_START))

	recent, err := chain.Conn.RpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("> GetRecentBlockhash: %v", err)
	}

	distri_ai.SetProgramID(chain.ProgramDistriID)
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			distri_ai.NewStartOrderInstruction(
				chain.ProgramDistriOrder,
				chain.Wallet.Wallet.PublicKey(),
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(chain.Wallet.Wallet.PublicKey()),
	)

	if err != nil {
		return "", fmt.Errorf("> solana.NewTransaction: %v", err)
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if chain.Wallet.Wallet.PublicKey().Equals(key) {
				return &chain.Wallet.Wallet.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("> tx.Sign: %v", err)
	}

	spew.Dump(tx)

	sig, err := chain.Conn.SendAndConfirmTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("> SendAndConfirmTransaction: %v", err.Error())
	}

	logs.Vital(fmt.Sprintf("%s completed : %v", pattern.TX_HASHRATE_MARKET_ORDER_START, sig))

	return sig, nil
}

func (chain WrapperDistri) OrderCompleted(orderPlacedMetadata pattern.OrderPlacedMetadata, isGPU bool) (string, error) {
	logs.Normal(fmt.Sprintf("Extrinsic : %v", pattern.TX_HASHRATE_MARKET_ORDER_COMPLETED))

	score, err := docker.RunScoreContainer(isGPU)
	if err != nil {
		return "", err
	}
	scoreUint8 := uint8(score)

	recent, err := chain.Conn.RpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	jsonData, err := json.Marshal(orderPlacedMetadata)
	if err != nil {
		return "", fmt.Errorf("error marshaling the struct to JSON: %v", err)
	}

	seller := chain.Wallet.Wallet.PublicKey()
	ecpc := solana.MustPublicKeyFromBase58(pattern.DIST_TOKEN_ID)
	sellerAta, _, err := solana.FindAssociatedTokenAddress(seller, ecpc)
	if err != nil {
		return "", fmt.Errorf("error finding associated token address: %v", err)
	}

	seedVault := utils.GenVault()
	vault, _, err := solana.FindProgramAddress(
		seedVault,
		chain.ProgramDistriID,
	)
	if err != nil {
		return "", fmt.Errorf("error finding program address: %v", err)
	}

	distri_ai.SetProgramID(chain.ProgramDistriID)
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			distri_ai.NewOrderCompletedInstruction(
				string(jsonData),
				scoreUint8,
				chain.ProgramDistriMachine,
				chain.ProgramDistriOrder,
				seller,
				sellerAta,
				vault,
				ecpc,
				solana.TokenProgramID,
				solana.SPLAssociatedTokenAccountProgramID,
				solana.SystemProgramID,
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(chain.Wallet.Wallet.PublicKey()),
	)

	if err != nil {
		return "", fmt.Errorf("error creating transaction: %v", err)
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if chain.Wallet.Wallet.PublicKey().Equals(key) {
				return &chain.Wallet.Wallet.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("error signing transaction: %v", err)
	}

	logs.Normal("=============== OrderCompleted Transaction ==================")
	spew.Dump(tx)

	sig, err := chain.Conn.SendAndConfirmTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("> SendAndConfirmTransaction: %v", err.Error())
	}

	logs.Vital(fmt.Sprintf("%s completed : %v", pattern.TX_HASHRATE_MARKET_ORDER_COMPLETED, sig))

	return sig, nil
}

func (chain WrapperDistri) OrderFailed(buyer solana.PublicKey, orderPlacedMetadata pattern.OrderPlacedMetadata) (string, error) {
	logs.Normal(fmt.Sprintf("Extrinsic : %v", pattern.TX_HASHRATE_MARKET_ORDER_FAILED))

	recent, err := chain.Conn.RpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}

	jsonData, err := json.Marshal(orderPlacedMetadata)
	if err != nil {
		return "", fmt.Errorf("> json.Marshal: %v", err.Error())
	}

	seller := chain.Wallet.Wallet.PublicKey()
	ecpc := solana.MustPublicKeyFromBase58(pattern.DIST_TOKEN_ID)
	buyerAta, _, err := solana.FindAssociatedTokenAddress(buyer, ecpc)
	if err != nil {
		return "", fmt.Errorf("> FindAssociatedTokenAddress: %v", err.Error())
	}

	seedVault := utils.GenVault()
	vault, _, err := solana.FindProgramAddress(
		seedVault,
		chain.ProgramDistriID,
	)
	if err != nil {
		return "", fmt.Errorf("> FindProgramAddress: %v", err.Error())
	}

	distri_ai.SetProgramID(chain.ProgramDistriID)
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			distri_ai.NewOrderFailedInstruction(
				string(jsonData),
				chain.ProgramDistriMachine,
				chain.ProgramDistriOrder,
				seller,
				buyerAta,
				vault,
				ecpc,
				solana.TokenProgramID,
				solana.SPLAssociatedTokenAccountProgramID,
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(chain.Wallet.Wallet.PublicKey()),
	)

	if err != nil {
		return "", fmt.Errorf("> NewOrderFailedInstruction: %v", err.Error())
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if chain.Wallet.Wallet.PublicKey().Equals(key) {
				return &chain.Wallet.Wallet.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("> tx.Sign: %v", err.Error())
	}

	spew.Dump(tx)

	sig, err := chain.Conn.SendAndConfirmTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("> SendAndConfirmTransaction: %v", err.Error())
	}

	logs.Vital(fmt.Sprintf("%s completed : %v", pattern.TX_HASHRATE_MARKET_ORDER_FAILED, sig))

	return sig, nil
}

func (chain WrapperDistri) GetMachine() (distri_ai.Machine, error) {

	var data distri_ai.Machine

	resp, err := chain.Conn.RpcClient.GetAccountInfo(
		context.TODO(),
		chain.ProgramDistriMachine,
	)
	if err != nil {
		return data, nil
	}

	borshDec := bin.NewBorshDecoder(resp.GetBinary())

	err = data.UnmarshalWithDecoder(borshDec)
	if err != nil {
		return data, fmt.Errorf("> UnmarshalWithDecoder: %v", err)
	}

	return data, nil
}

func (chain WrapperDistri) GetOrder() (distri_ai.Order, error) {

	var data distri_ai.Order

	resp, err := chain.Conn.RpcClient.GetAccountInfo(
		context.TODO(),
		chain.ProgramDistriOrder,
	)
	if err != nil {
		return data, nil
	}

	borshDec := bin.NewBorshDecoder(resp.GetBinary())

	err = data.UnmarshalWithDecoder(borshDec)
	if err != nil {
		return data, fmt.Errorf("error unmarshaling data: %v", err)
	}

	return data, nil
}

func (chain WrapperDistri) SubmitTask(
	taskUuid pattern.TaskUUID,
	machineUUID pattern.MachineUUID,
	period uint32,
	taskMetadata pattern.TaskMetadata) (string, error) {
	logs.Normal(fmt.Sprintf("Extrinsic : %v", pattern.TX_HASHRATE_MARKET_SUBMIT_TASK))

	recent, err := chain.Conn.RpcClient.GetRecentBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		return "", fmt.Errorf("error getting recent blockhash: %v", err)
	}

	jsonData, err := json.Marshal(taskMetadata)
	if err != nil {
		return "", fmt.Errorf("error marshaling the struct to JSON: %v", err)
	}

	programID := solana.MustPublicKeyFromBase58(pattern.PROGRAM_DISTRI_ID)
	seedTask := utils.GenTask(chain.Wallet.Wallet.PublicKey(), taskUuid)
	task, _, _ := solana.FindProgramAddress(
		seedTask,
		programID,
	)
	seedReward := utils.GenReward()
	reward, _, _ := solana.FindProgramAddress(
		seedReward,
		programID,
	)
	seedRewardMachine := utils.GenRewardMachine(chain.Wallet.Wallet.PublicKey(), machineUUID)
	rewardMachine, _, _ := solana.FindProgramAddress(
		seedRewardMachine,
		programID,
	)

	distri_ai.SetProgramID(chain.ProgramDistriID)
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			distri_ai.NewSubmitTaskInstruction(
				taskUuid,
				utils.CurrentPeriod(),
				string(jsonData),
				chain.ProgramDistriMachine,
				task,
				reward,
				rewardMachine,
				chain.Wallet.Wallet.PublicKey(),
				solana.SystemProgramID,
			).Build(),
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(chain.Wallet.Wallet.PublicKey()),
	)

	if err != nil {
		return "", fmt.Errorf("error creating transaction: %v", err)
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if chain.Wallet.Wallet.PublicKey().Equals(key) {
				return &chain.Wallet.Wallet.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return "", fmt.Errorf("error signing transaction: %v", err)
	}

	spew.Dump(tx)

	sig, err := chain.Conn.SendAndConfirmTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("> SendAndConfirmTransaction: %v", err.Error())
	}

	logs.Vital(fmt.Sprintf("%s completed : %v", pattern.TX_HASHRATE_MARKET_SUBMIT_TASK, sig))

	return sig, nil
}

func NewDistriWrapper(info *chain.InfoChain) *WrapperDistri {
	return &WrapperDistri{info}
}
